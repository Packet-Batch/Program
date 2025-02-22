package network

import (
	"fmt"
	"strconv"
	"strings"
)

const IPPROTO_ICMP = 1
const IPPROTO_TCP = 6
const IPPROTO_UDP = 17

func GetMacOfInterface(dev string) [6]byte {
	var macAddr [6]byte

	return macAddr
}

func GetGatewayMacAddr() [6]byte {
	var macAddr [6]byte

	return macAddr
}

func MacAddrStrToArr(str string) ([6]byte, error) {
	var macAddr [6]byte
	parts := strings.Split(str, ":")

	if len(parts) < 6 {
		return macAddr, fmt.Errorf("mac address specified, but malformed")
	}

	for i, v := range parts {
		bVal, err := strconv.ParseUint(v, 16, 8)

		if err != nil {
			return macAddr, fmt.Errorf("failed to convert byte to uint at index %d: %v", i, err)
		}

		macAddr[i] = byte(bVal)
	}

	return macAddr, nil
}

func GetProtocolIdByStr(proto string) int {
	protoLower := strings.ToLower(proto)

	switch protoLower {
	case "tcp":
		return IPPROTO_TCP

	case "icmp":
		return IPPROTO_ICMP

	default:
		return IPPROTO_UDP
	}
}

func GetProtocolStrById(id int) string {
	switch id {
	case IPPROTO_TCP:
		return "tcp"

	case IPPROTO_ICMP:
		return "icmp"

	case IPPROTO_UDP:
		return "udp"
	}

	return ""
}
