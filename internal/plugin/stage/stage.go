package stage

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/units"
	log "github.com/hashicorp/go-hclog"
	"github.com/oclaussen/go-gimme/ssh"
	"github.com/pkg/errors"
	coreapi "github.com/wabenet/dodo-core/api/v1alpha4"
	coreconfig "github.com/wabenet/dodo-core/pkg/config"
	"github.com/wabenet/dodo-core/pkg/plugin"
	"github.com/wabenet/dodo-core/pkg/plugin/builder"
	"github.com/wabenet/dodo-core/pkg/plugin/runtime"
	"github.com/wabenet/dodo-stage-virtualbox/internal/config"
	"github.com/wabenet/dodo-stage-virtualbox/pkg/virtualbox"
	api "github.com/wabenet/dodo-stage/api/v1alpha2"
	"github.com/wabenet/dodo-stage/pkg/box"
	"github.com/wabenet/dodo-stage/pkg/integrations/ova"
	"github.com/wabenet/dodo-stage/pkg/plugin/stage"
	"github.com/wabenet/dodo-stage/pkg/proxy"
	"github.com/wabenet/dodo-stage/pkg/stagehand"
	"github.com/wabenet/dodo-stage/pkg/stagehand/installer"
)

const (
	name        = "virtualbox"
	defaultPort = 20257 // TODO: this is not the place for this
)

var _ stage.Stage = &Stage{}

type Stage struct {
	proxyClient *proxy.Client
}

func New() *Stage {
	return &Stage{}
}

func (vbox *Stage) Type() plugin.Type {
	return stage.Type
}

func (vbox *Stage) PluginInfo() *coreapi.PluginInfo {
	return &coreapi.PluginInfo{
		Name: &coreapi.PluginName{Name: "virtualbox", Type: stage.Type.String()},
	}
}

func (vbox *Stage) Init() (plugin.PluginConfig, error) {
	return map[string]string{}, nil
}

func (vbox *Stage) Cleanup() {
	if vbox.proxyClient != nil {
		vbox.proxyClient.Close()
		vbox.proxyClient = nil
	}
}

func (vbox *Stage) getProxyClient(name string) (*proxy.Client, error) {
	if vbox.proxyClient != nil {
		return vbox.proxyClient, nil
	}

	url, err := vbox.GetURL(name)
	if err != nil {
		return nil, err
	}

	pc, err := proxy.NewClient(&proxy.Config{
		Address:  url,
		CAFile:   filepath.Join(storagePath(name), "ca.pem"),
		CertFile: filepath.Join(storagePath(name), "client.pem"),
		KeyFile:  filepath.Join(storagePath(name), "client-key.pem"),
	})
	if err != nil {
		return nil, err
	}

	vbox.proxyClient = pc

	return pc, nil
}

func (vbox *Stage) GetStage(name string) (*api.GetStageResponse, error) {
	result := &api.GetStageResponse{Name: name}

	exist, err := vbox.Exist(name)
	if err != nil {
		return nil, err
	}

	result.Exist = exist
	if !result.Exist {
		return result, nil
	}

	available, err := vbox.Available(name)
	if err != nil {
		return nil, err
	}

	result.Available = available
	if !result.Available {
		return result, nil
	}

	sshOpts, err := vbox.GetSSHOptions(name)
	if err != nil {
		return nil, err
	}

	result.SshOptions = sshOpts

	return result, nil
}

func (vbox *Stage) CreateStage(conf *api.Stage) error {
	stages, err := config.GetAllStages(coreconfig.GetConfigFiles()...)
	if err != nil {
		return err
	}

	options := stages[conf.Name].Options
	vm := &virtualbox.VM{Name: conf.Name}

	if err := os.MkdirAll(storagePath(conf.Name), 0700); err != nil {
		return err
	}

	log.L().Info("creating SSH key...")
	if _, err := ssh.GimmeKeyPair(filepath.Join(storagePath(conf.Name), "id_rsa")); err != nil {
		return errors.Wrap(err, "could not generate SSH key")
	}

	b, err := box.Load(conf.Box, "virtualbox")
	if err != nil {
		return errors.Wrap(err, "could not load box")
	}
	if err := b.Download(); err != nil {
		return errors.Wrap(err, "could not download box")
	}

	sshOpts, err := b.GetSSHOptions()
	if err != nil {
		return err
	}

	if err := saveState(conf.Name, &State{
		Username:       sshOpts.Username,
		PrivateKeyFile: sshOpts.PrivateKeyFile,
	}); err != nil {
		return err
	}

	boxFile := filepath.Join(b.Path(), "box.ovf")
	ovf, err := ova.ReadOVF(boxFile)
	if err != nil {
		return err
	}

	importArgs := []string{boxFile, "--vsys", "0", "--vmname", vm.Name, "--basefolder", storagePath(conf.Name)}
	for _, item := range ovf.VirtualSystem.VirtualHardware.Items {
		switch item.ResourceType {
		case ova.TypeCPU:
			if cpu := conf.Resources.Cpu; cpu > 0 {
				importArgs = append(importArgs, "--vsys", "0", "--cpus", fmt.Sprintf("%d", cpu))
			}
		case ova.TypeMemory:
			if memory := conf.Resources.Memory; memory > 0 {
				memory = memory / int64(units.Megabyte)
				importArgs = append(importArgs, "--vsys", "0", "--memory", fmt.Sprintf("%d", memory))
			}
		}
	}

	if err := vm.Import(importArgs...); err != nil {
		return errors.Wrap(err, "could not import VM")
	}

	if err := vm.Modify(
		"--firmware", "bios",
		"--bioslogofadein", "off",
		"--bioslogofadeout", "off",
		"--bioslogodisplaytime", "0",
		"--biosbootmenu", "disabled",
		"--acpi", "on",
		"--ioapic", "on",
		"--rtcuseutc", "on",
		"--natdnshostresolver1", "off",
		"--natdnsproxy1", "on",
		"--cpuhotplug", "off",
		"--pae", "on",
		"--hpet", "on",
		"--hwvirtex", "on",
		"--nestedpaging", "on",
		"--largepages", "on",
		"--vtxvpid", "on",
		"--accelerate3d", "off",
	); err != nil {
		return errors.Wrap(err, "could not configure general VM settings")
	}

	if err := vm.Modify(
		"--nic1", "nat",
		"--nictype1", "82540EM",
		"--cableconnected1", "on",
	); err != nil {
		return errors.Wrap(err, "could not create nat controller")
	}

	if len(options.Modify) > 0 {
		if err := vm.Modify(options.Modify...); err != nil {
			return err
		}
	}

	sataController, err := vm.GetStorageController(virtualbox.SATA)
	if err != nil {
		return err
	}

	numDisks := len(sataController.Disks)
	for index, volume := range conf.Resources.Volumes {
		disk := virtualbox.Disk{
			Path: filepath.Join(persistPath(conf.Name), fmt.Sprintf("disk-%d.vmdk", index)),
			Size: volume.Size,
		}
		if err := disk.Create(); err != nil {
			return err
		}
		if err := sataController.AttachDisk(numDisks+index, &disk); err != nil {
			return err
		}
	}

	for index, usb := range conf.Resources.UsbFilters {
		filter := virtualbox.USBFilter{
			VMName:    vm.Name,
			Index:     index,
			Name:      usb.Name,
			VendorID:  usb.VendorId,
			ProductID: usb.ProductId,
		}
		if err := filter.Create(); err != nil {
			return err
		}
	}

	return nil
}

func (vbox *Stage) StartStage(name string) error {
	vm := &virtualbox.VM{Name: name}

	running, err := vbox.Available(name)
	if err != nil {
		return err
	}

	if running {
		return errors.New("VM is already running")
	}
	log.L().Info("starting VM...")

	log.L().Info("configure network...")
	if err := vbox.SetupHostOnlyNetwork(name, "192.168.99.1/24"); err != nil {
		return errors.Wrap(err, "could not set up host-only network")
	}

	sshForwarding := vm.NewPortForwarding("ssh")
	sshForwarding.GuestPort = 22
	if err := sshForwarding.Create(); err != nil {
		return errors.Wrap(err, "could not configure port forwarding")
	}

	if err := vm.Start(); err != nil {
		return errors.Wrap(err, "could not start VM")
	}

	log.L().Info("waiting for SSH...")
	if err = await(func() (bool, error) {
		return vbox.isSSHAvailable(name)
	}); err != nil {
		return err
	}

	log.L().Info("VM is running")
	return nil
}

func (vbox *Stage) ProvisionStage(name string) error {
	stages, err := config.GetAllStages(coreconfig.GetConfigFiles()...)
	if err != nil {
		return err
	}

	options := stages[name].Options
	vm := &virtualbox.VM{Name: name}

	sshOpts, err := vbox.GetSSHOptions(name)
	if err != nil {
		return err
	}

	publicKey, err := ioutil.ReadFile(filepath.Join(storagePath(name), "id_rsa.pub"))
	if err != nil {
		return err
	}

	inst := installer.SSHInstaller{
		DownloadUrl: options.StagehandURL,
		SSHOptions:  sshOpts,
	}

	provisionConfig := &stagehand.Config{
		Hostname:          vm.Name,
		DefaultUser:       sshOpts.Username,
		AuthorizedSSHKeys: []string{string(publicKey)},
		Script:            options.Provision,
	}

	result, err := inst.Install(provisionConfig)
	if err != nil {
		return err
	}

	state, err := loadState(name)
	if err != nil {
		return err
	}

	state.IPAddress = result.IPAddress
	state.PrivateKeyFile = filepath.Join(storagePath(name), "id_rsa")
	if err := saveState(name, state); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(storagePath(name), "ca.pem"), []byte(result.CA), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(storagePath(name), "client.pem"), []byte(result.ClientCert), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(storagePath(name), "client-key.pem"), []byte(result.ClientKey), 0600); err != nil {
		return err
	}

	pemData, _ := pem.Decode([]byte(result.CA))
	caCert, err := x509.ParseCertificate(pemData.Bytes)
	if err != nil {
		return err
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)

	keyPair, err := tls.X509KeyPair([]byte(result.ClientCert), []byte(result.ClientKey))
	if err != nil {
		return err
	}

	boxUrl, err := vbox.GetURL(name)
	if err != nil {
		return err
	}
	parsed, err := url.Parse(boxUrl)
	if err != nil {
		return errors.Wrap(err, "could not parse URL")
	}

	if _, err = tls.DialWithDialer(
		&net.Dialer{Timeout: 20 * time.Second},
		"tcp",
		parsed.Host,
		&tls.Config{
			RootCAs:            certPool,
			InsecureSkipVerify: false,
			Certificates:       []tls.Certificate{keyPair},
		},
	); err != nil {
		return err
	}

	log.L().Info("VM is provisioned")
	return nil
}

func (vbox *Stage) StopStage(name string) error {
	vm := &virtualbox.VM{Name: name}
	log.L().Info("stopping VM...")

	available, err := vbox.Available(name)
	if err != nil {
		return err
	}
	if !available {
		log.L().Info("VM is already stopped")
		return nil
	}

	if err := vm.Stop(false); err != nil {
		return err
	}

	if err := await(func() (bool, error) {
		available, err := vbox.Available(name)
		return !available, err
	}); err != nil {
		return err
	}

	return errors.New("VM did not stop successfully")
}

func (vbox *Stage) DeleteStage(name string, force bool, volumes bool) error {
	vm := &virtualbox.VM{Name: name}

	exist, err := vbox.Exist(name)
	if err != nil {
		if !force {
			return err
		}
	}

	if !exist && !force {
		log.L().Info("VM does not exist")
		return nil
	}

	log.L().Info("removing VM...")

	running, err := vbox.Available(name)
	if err != nil {
		if !force {
			return err
		}
	}

	if running {
		if err := vm.Stop(true); err != nil {
			if !force {
				return err
			}
		}
	}

	if err = vm.Delete(); err != nil {
		if !force {
			return err
		}
	}

	if err := os.RemoveAll(storagePath(name)); err != nil {
		if !force {
			return err
		}
	}

	if volumes {
		if err := os.RemoveAll(persistPath(name)); err != nil {
			if !force {
				return err
			}
		}
	}

	log.L().Info("removed VM")
	return nil
}

func (vbox *Stage) Exist(name string) (bool, error) {
	vm := &virtualbox.VM{Name: name}

	if _, err := os.Stat(storagePath(name)); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	_, err := vm.Info()
	return err == nil, nil
}

func (vbox *Stage) Available(name string) (bool, error) {
	vm := &virtualbox.VM{Name: name}

	info, err := vm.Info()
	if err != nil {
		return false, err
	}
	state, ok := info["VMState"]
	return ok && state == "running", nil
}

func (vbox *Stage) GetURL(name string) (string, error) {
	state, err := loadState(name)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:%d", state.IPAddress, defaultPort), nil
}

func (vbox *Stage) GetSSHOptions(name string) (*api.SSHOptions, error) {
	vm := &virtualbox.VM{Name: name}

	portForwardings, err := vm.ListPortForwardings()
	if err != nil {
		return nil, err
	}

	port := 0
	for _, forward := range portForwardings {
		if forward.Name == "ssh" {
			port = forward.HostPort
			break
		}
	}
	if port == 0 {
		return nil, errors.New("no port forwarding matching ssh port found")
	}

	state, err := loadState(name)
	if err != nil {
		return nil, err
	}

	return &api.SSHOptions{
		Hostname:       "127.0.0.1",
		Port:           int32(port),
		Username:       state.Username,
		PrivateKeyFile: state.PrivateKeyFile,
	}, nil
}

func (vbox *Stage) GetContainerRuntime(name string) (runtime.ContainerRuntime, error) {
	pc, err := vbox.getProxyClient(name)
	if err != nil {
		return nil, err
	}

	return pc.ContainerRuntime, nil
}

func (vbox *Stage) GetImageBuilder(name string) (builder.ImageBuilder, error) {
	pc, err := vbox.getProxyClient(name)
	if err != nil {
		return nil, err
	}

	return pc.ImageBuilder, nil
}

func storagePath(name string) string {
	return filepath.Join(coreconfig.GetAppDir(), "stages", name)
}

func persistPath(name string) string {
	return filepath.Join(coreconfig.GetAppDir(), "persist", name)
}
