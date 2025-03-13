package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/google/gopacket/layers"
)

func GetMacOfInterface(dev string) ([]byte, error) {
	var macAddr []byte

	addrPath := fmt.Sprintf("/sys/class/net/%s/address", dev)

	contents, err := os.ReadFile(addrPath)

	if err != nil {
		return macAddr, err
	}

	return MacAddrStrToArr(string(contents))
}

// getDefaultGateway retrieves the IP address of the default gateway
func getDefaultGateway() (string, error) {
	cmd := exec.Command("sh", "-c", "ip -4 route list 0/0 | awk '{print $3}'")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get default gateway: %v", err)
	}

	gatewayIP := strings.TrimSpace(out.String())
	if gatewayIP == "" {
		return "", fmt.Errorf("no default gateway found")
	}
	return gatewayIP, nil
}

// GetGatewayMacAddr retrieves the MAC address of the default gateway
func GetGatewayMacAddr() ([]byte, error) {
	gatewayIP, err := getDefaultGateway()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf("ip neigh | grep -m1 '%s ' | awk '{print $5}'", gatewayIP))
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get gateway MAC: %v", err)
	}

	macStr := strings.TrimSpace(out.String())

	if macStr == "" {
		return nil, fmt.Errorf("could not retrieve gateway MAC address")
	}

	return MacAddrStrToArr(macStr)
}

func MacAddrStrToArr(str string) ([]byte, error) {
	var macAddr []byte

	str = strings.TrimSpace(str)

	parts := strings.Split(str, ":")

	if len(parts) < 6 {
		return macAddr, fmt.Errorf("mac address specified, but malformed")
	}

	for i, v := range parts {
		bVal, err := strconv.ParseUint(v, 16, 8)

		if err != nil {
			return macAddr, fmt.Errorf("failed to convert byte to uint at index %d: %v", i, err)
		}

		macAddr = append(macAddr, byte(bVal))
	}

	return macAddr, nil
}

func GetProtocolIdByStr(proto string) (layers.IPProtocol, error) {
	protoLower := strings.ToLower(proto)

	switch protoLower {
	case "tcp":
		return layers.IPProtocolTCP, nil

	case "icmp":
		return layers.IPProtocolICMPv4, nil

	case "udp":
		return layers.IPProtocolUDP, nil
	}

	return layers.IPProtocolUDP, fmt.Errorf("invalid protocol")
}

func GetProtocolStrById(proto layers.IPProtocol) (string, error) {
	switch proto {
	case layers.IPProtocolTCP:
		return "tcp", nil

	case layers.IPProtocolICMPv4:
		return "icmp", nil

	case layers.IPProtocolUDP:
		return "udp", nil
	}

	return "", fmt.Errorf("protocol not found")
}

func GetIpFromRange(ipRange string, rng *rand.Rand) (uint32, error) {
	var err error

	// Split the IP range and retrieve CIDR if available (otherwise use 32).
	parts := strings.Split(ipRange, "/")

	cidr := 32

	if len(parts) == 2 {
		cidr, err = strconv.Atoi(parts[1])

		if err != nil {
			return 0, fmt.Errorf("failed to parse CIDR range '%s': %v", parts[1], err)
		}
	}

	ipStr := parts[0]

	// Parse the network IP address to a 32-bit integer.
	ip := net.ParseIP(ipStr).To4()

	if ip == nil {
		return 0, fmt.Errorf("failed to parse network IP '%s': %v", ipStr, err)
	}

	ipAddr := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])

	// Generate the subnet mask based on CIDR.
	mask := uint32(1<<uint(32-cidr)) - 1

	// Generate random number.
	randNum := uint32(rng.Int31())

	// Generate random IP.
	randIp := (ipAddr &^ mask) | (randNum & mask)

	return randIp, nil
}

func U32ToNetIp(ip uint32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, ip)
	return net.IP(b)
}
