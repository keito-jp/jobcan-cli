package dashen

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/gopacket/pcap"
)

type Skip struct {
	Reason string
}

func (err Skip) Error() string {
	return fmt.Sprintf("skip because %s", err.Reason)
}

func Filter(iface pcap.Interface) error {
	if strings.Contains(iface.Name, "docker") {
		return &Skip{"docker"}
	}
	if strings.HasPrefix(iface.Name, "br-") {
		return &Skip{"bridge"}
	}
	if strings.HasPrefix(iface.Name, "usb") {
		return &Skip{"usb"}
	}
	if strings.HasPrefix(iface.Name, "bluetooth") {
		return &Skip{"bluetooth"}
	}

	as := iface.Addresses
	ok := false
	for _, a := range as {
		ip4 := a.IP.To4()
		if ip4 == nil {
			continue
		}
		if ip4[0] == 127 {
			continue
		}
		ok = true
		break
	}
	if !ok {
		return &Skip{"bad address"}
	}
	return nil
}

func IsEquals(i1, i2 interface{}) bool {
	return reflect.ValueOf(i1).Pointer() == reflect.ValueOf(i2).Pointer()
}
