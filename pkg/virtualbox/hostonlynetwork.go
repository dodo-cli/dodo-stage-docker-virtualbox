package virtualbox

import (
	"encoding/binary"
	"net"
	"regexp"
	"strings"
)

var empty interface{}

type HostOnlyNetwork struct {
	Name        string
	LowerIP     net.IP
	UpperIP     net.IP
	NetMask     net.IPMask
	NetworkName string
}

func NewHostOnlyNetwork(name string, cidr string) (*HostOnlyNetwork, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	mask := binary.BigEndian.Uint32(network.Mask)
	rangeStart := binary.BigEndian.Uint32(network.IP)
	rangeEnd := (rangeStart & mask) | (mask ^ 0xffffffff)

	lowerIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(lowerIP, rangeStart)

	upperIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(upperIP, rangeEnd)

	return &HostOnlyNetwork{
		Name:    name,
		LowerIP: lowerIP,
		UpperIP: upperIP,
		NetMask: network.Mask,
	}, nil
}

func (network HostOnlyNetwork) Equal(other HostOnlyNetwork) bool {
	if !network.LowerIP.Equal(other.LowerIP) {
		return false
	}

	if !network.UpperIP.Equal(other.UpperIP) {
		return false
	}

	if network.NetMask.String() != other.NetMask.String() {
		if network.NetMask.String() != "0f000000" {
			return false
		}
	}

	return true
}

func (network *HostOnlyNetwork) Create() error {
	existingNetworks, err := ListHostOnlyNetworks()
	if err != nil {
		return err
	}

	networkNames := map[string]interface{}{}
	for _, n := range existingNetworks {
		if n.Equal(*network) {
			network.Name = n.Name
			network.NetworkName = n.NetworkName
			return nil
		}
		networkNames[n.NetworkName] = empty
	}

	_, err = vbm(
		"hostonlynet", "add",
		"--name", network.Name,
		"--netmask", net.IP(network.NetMask).String(),
		"--lower-ip", network.LowerIP.String(),
		"--upper-ip", network.UpperIP.String(),
	)

	return err
}

func (network *HostOnlyNetwork) ConnectVM(vm *VM) error {
	return vm.Modify(
		"--nic2", "hostonlynet",
		"--nictype2", "82540EM",
		"--nicpromisc2", "deny",
		"--host-only-net2", network.Name,
		"--cableconnected2", "on",
	)
}

func ListHostOnlyNetworks() ([]*HostOnlyNetwork, error) {
	stdout, err := vbm("list", "hostonlynets")
	if err != nil {
		return nil, err
	}

	result := []*HostOnlyNetwork{}
	current := &HostOnlyNetwork{}
	re := regexp.MustCompile(`(.+):\s+(.*)`)
	for _, line := range strings.Split(stdout, "\n") {
		if line == "" {
			continue
		}

		groups := re.FindStringSubmatch(line)
		if groups == nil {
			continue
		}

		switch groups[1] {
		case "Name":
			current = &HostOnlyNetwork{Name: groups[2]}
			result = append(result, current)
		case "LowerIP":
			current.LowerIP = net.ParseIP(groups[2])
		case "UpperIP":
			current.UpperIP = net.ParseIP(groups[2])
		case "NetworkMask":
			current.NetMask = parseIPv4Mask(groups[2])
		case "VBoxNetworkName":
			current.NetworkName = groups[2]
		}
	}

	return result, nil
}
