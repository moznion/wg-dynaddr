package internal

import (
	"github.com/vishvananda/netlink"
)

// NetworkInterfaceAddressReloader defines the methods to reload the interface's addresses.
type NetworkInterfaceAddressReloader interface {
	// Reload reloads address information by given addresses.
	Reload(addresses []string) error
}

// DefaultNetworkInterfaceAddressReloader is a default implementation of NetworkInterfaceAddressReloader.
type DefaultNetworkInterfaceAddressReloader struct {
	deviceName string
}

// NewDefaultNetworkInterfaceAddressReloader makes an instance of DefaultDNSRecordWatcher.
func NewDefaultNetworkInterfaceAddressReloader(deviceName string) (*DefaultNetworkInterfaceAddressReloader, error) {
	return &DefaultNetworkInterfaceAddressReloader{
		deviceName: deviceName,
	}, nil
}

// See NetworkInterfaceAddressReloader.Reload.
func (r *DefaultNetworkInterfaceAddressReloader) Reload(addresses []string) error {
	link, err := netlink.LinkByName(r.deviceName)
	if err != nil {
		return err
	}

	var addrModifier = netlink.AddrReplace // at the first time, it has to *replace* the address
	for _, address := range addresses {
		addr, err := netlink.ParseAddr(address)
		if err != nil {
			return err
		}
		err = addrModifier(link, addr)
		if err != nil {
			return err
		}

		addrModifier = netlink.AddrAdd
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return err
	}

	return nil
}
