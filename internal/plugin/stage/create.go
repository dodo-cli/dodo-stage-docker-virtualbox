package stage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/units"
	log "github.com/hashicorp/go-hclog"
	coreconfig "github.com/wabenet/dodo-core/pkg/config"
	"github.com/wabenet/dodo-stage-virtualbox/internal/config"
	"github.com/wabenet/dodo-stage-virtualbox/pkg/virtualbox"
	"github.com/wabenet/dodo-stage/pkg/box"
	"github.com/wabenet/dodo-stage/pkg/util/ova"
	"github.com/wabenet/dodo-stage/pkg/util/ssh"
	"github.com/wabenet/dodo-stage/pkg/util/state"
)

func (s *Stage) CreateStage(name string) error {
	stages, err := config.GetAllStages(coreconfig.GetConfigFiles()...)
	if err != nil {
		return err
	}

	stage := stages[name]
	vm := &virtualbox.VM{Name: name}

	if err := os.MkdirAll(storagePath(name), 0700); err != nil {
		return err
	}

	log.L().Info("creating SSH key...")
	if _, err := ssh.NewKeyPair(filepath.Join(storagePath(name), "id_rsa")); err != nil {
		return fmt.Errorf("could not generate SSH key: %q", err)
	}

	b, err := box.Load(stage.Box, "virtualbox")
	if err != nil {
		return fmt.Errorf("could not load box: %q", err)
	}
	if err := b.Download(); err != nil {
		return fmt.Errorf("could not download box: %q", err)
	}

	sshOpts, err := b.GetSSHOptions()
	if err != nil {
		return err
	}

	if err := state.Save(name, &State{
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

	importArgs := []string{boxFile, "--vsys", "0", "--vmname", vm.Name, "--basefolder", storagePath(name)}
	for _, item := range ovf.VirtualSystem.VirtualHardware.Items {
		switch item.ResourceType {
		case ova.TypeCPU:
			if cpu := stage.Resources.Cpu; cpu > 0 {
				importArgs = append(importArgs, "--vsys", "0", "--cpus", fmt.Sprintf("%d", cpu))
			}
		case ova.TypeMemory:
			if memory := stage.Resources.Memory; memory > 0 {
				memory = memory / int64(units.Megabyte)
				importArgs = append(importArgs, "--vsys", "0", "--memory", fmt.Sprintf("%d", memory))
			}
		}
	}

	if err := vm.Import(importArgs...); err != nil {
		return fmt.Errorf("could not import VM: %q", err)
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
		return fmt.Errorf("could not configure general VM settings: %q", err)
	}

	if err := vm.Modify(
		"--nic1", "nat",
		"--nictype1", "82540EM",
		"--cableconnected1", "on",
	); err != nil {
		return fmt.Errorf("could not create nat controller: %q", err)
	}

	if len(stage.Options) > 0 {
		if err := vm.Modify(stage.Options...); err != nil {
			return err
		}
	}

	sataController, err := vm.GetStorageController(virtualbox.SATA)
	if err != nil {
		return err
	}

	numDisks := len(sataController.Disks)
	for index, volume := range stage.Resources.Volumes {
		disk := virtualbox.Disk{
			Path: filepath.Join(persistPath(name), fmt.Sprintf("disk-%d.vmdk", index)),
			Size: volume.Size,
		}
		if err := disk.Create(); err != nil {
			return err
		}
		if err := sataController.AttachDisk(numDisks+index, &disk); err != nil {
			return err
		}
	}

	for index, usb := range stage.Resources.UsbFilters {
		filter := virtualbox.USBFilter{
			VMName:    vm.Name,
			Index:     index,
			Name:      usb.Name,
			VendorID:  usb.VendorID,
			ProductID: usb.ProductID,
		}
		if err := filter.Create(); err != nil {
			return err
		}
	}

	return nil
}
