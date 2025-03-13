package sequence

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/Packet-Batch/Program/internal/cli"
	"github.com/Packet-Batch/Program/internal/config"
	"github.com/Packet-Batch/Program/internal/network"
	"github.com/Packet-Batch/Program/internal/tech"
	"github.com/Packet-Batch/Program/internal/tech/afpacket"
	"github.com/Packet-Batch/Program/internal/tech/afxdp"
	"github.com/Packet-Batch/Program/internal/tech/dpdk"
	"github.com/Packet-Batch/Program/internal/utils"
	"github.com/google/gopacket"

	"github.com/google/gopacket/layers"
	_ "github.com/google/gopacket/layers"
)

func ProcessSeq(cfg *config.Config, cli *cli.Cli, seq *config.Sequence) error {
	var err error

	// Load tech.
	t, err := tech.Load(seq.Tech)

	if err != nil {
		return fmt.Errorf("failed to load tech '%s': %v", seq.Tech, err)
	}

	dev := seq.Interface

	if len(dev) < 1 {
		dev = cfg.Interface

		if len(dev) < 1 {
			return fmt.Errorf("no interface set in sequence or config")
		}
	}

	// Get thread count.
	threads := seq.Threads

	if threads < 1 {
		threads = uint8(utils.GetCpuCount())
	}

	if threads < 1 {
		return fmt.Errorf("threads below 1")
	}

	// Retrieve source MAC address.
	srcMac := []byte{}

	if len(seq.Eth.SrcMac) > 0 {
		srcMac, err = network.MacAddrStrToArr(seq.Eth.SrcMac)

		if err != nil {
			return fmt.Errorf("failed to parse source MAC address: %v", err)
		}
	} else {
		srcMac, err = network.GetMacOfInterface(dev)

		if err != nil {
			return fmt.Errorf("failed to retrieve source MAC address of interface '%s': %v", dev, err)
		}
	}

	// Retrieve destination MAC address.
	dstMac := []byte{}

	if len(seq.Eth.DstMac) > 0 {
		dstMac, err = network.MacAddrStrToArr(seq.Eth.DstMac)

		if err != nil {
			return fmt.Errorf("failed to parse destination MAC address: %v", err)
		}
	} else {
		dstMac, err = network.GetGatewayMacAddr()

		if err != nil {
			return fmt.Errorf("failed to retrieve default MAC address: %v", err)
		}
	}

	cfg.DebugMsg(3, "[SEQ] Using src MAC => %x:%x:%x:%x:%x:%x", srcMac[0], srcMac[1], srcMac[2], srcMac[3], srcMac[4], srcMac[5])
	cfg.DebugMsg(3, "[SEQ] Using dst MAC => %x:%x:%x:%x:%x:%x", dstMac[0], dstMac[1], dstMac[2], dstMac[3], dstMac[4], dstMac[5])

	cAfxdp, _ := t.(*afxdp.Context)
	cAfpacket, _ := t.(*afpacket.Context)
	cDpdk, _ := t.(*dpdk.Context)

	// Setup tech.
	switch seq.Tech {
	case "af_packet":
		err := cAfpacket.Setup(dev, seq.Tcp.UseCookedSocket, int(seq.Threads))

		if err != nil {
			return fmt.Errorf("failed to setup AF_PACKET sequence: %v", err)
		}

	case "dpdk":
		err := cDpdk.Setup(dev, int(seq.Threads))

		if err != nil {
			return fmt.Errorf("failed to setup DPDK sequence: %v", err)
		}

	default:
		err := cAfxdp.Setup(dev, cli.AfXdp.Queue, cli.AfXdp.NeedWakeup, cli.AfXdp.SharedUmem, cli.AfXdp.ForceSkb, cli.AfXdp.ZeroCopy, int(threads))

		if err != nil {
			return fmt.Errorf("failed to setup AF_XDP sequence: %v", err)
		}
	}

	// Create random seed.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	lays := []gopacket.SerializableLayer{}

	// Ethernet header.
	eth := layers.Ethernet{
		EthernetType: layers.EthernetTypeIPv4,
		SrcMAC:       srcMac,
		DstMAC:       dstMac,
	}

	// IP header.
	iph := layers.IPv4{
		Version:    4,
		IHL:        5,
		FragOffset: 0,
		TOS:        seq.Ip4.Tos,
	}

	// Check for static IP fields.
	if seq.Ip4.MinTtl == seq.Ip4.MaxTtl {
		iph.TTL = seq.Ip4.MinTtl
	}

	if seq.Ip4.MinId == seq.Ip4.MaxId {
		iph.Id = seq.Ip4.MinId
	}

	if len(seq.Ip4.SrcIp) > 0 {
		srcIp, err := net.ResolveIPAddr("ip", seq.Ip4.SrcIp)

		if err != nil {
			return fmt.Errorf("failed to resolve source Ipv4 address '%s': %v", seq.Ip4.SrcIp, err)
		}

		iph.SrcIP = srcIp.IP
	}

	// Check and fill destination IP address.
	if len(seq.Ip4.DstIp) < 1 {
		return fmt.Errorf("no destination address set")
	}

	dstIp, err := net.ResolveIPAddr("ip", seq.Ip4.DstIp)

	if err != nil {
		return fmt.Errorf("failed to resolve destination IP address'%s': %v", seq.Ip4.DstIp, err)
	}

	iph.DstIP = dstIp.IP

	// Check and set IP protocol.
	proto, err := network.GetProtocolIdByStr(seq.Ip4.Protocol)

	if err != nil {
		return fmt.Errorf("failed to find protocol by string: %v", err)
	}

	iph.Protocol = proto

	lays = append(lays, &eth, &iph)

	// Handle layer-4 protocols.
	tcph := layers.TCP{}
	udph := layers.UDP{}
	icmph := layers.ICMPv4{}

	switch proto {
	case layers.IPProtocolTCP:
		if seq.Tcp.SrcPort > 0 {
			tcph.SrcPort = layers.TCPPort(seq.Tcp.SrcPort)
		}

		if seq.Tcp.DstPort > 0 {
			tcph.DstPort = layers.TCPPort(seq.Tcp.DstPort)
		}

		lays = append(lays, &tcph)

	case layers.IPProtocolUDP:
		if seq.Udp.SrcPort > 0 {
			udph.SrcPort = layers.UDPPort(seq.Tcp.SrcPort)
		}

		if seq.Udp.DstPort > 0 {
			udph.DstPort = layers.UDPPort(seq.Tcp.DstPort)
		}

		lays = append(lays, &udph)

	case layers.IPProtocolICMPv4:
		icmph.TypeCode = layers.ICMPv4TypeCode((uint16(seq.Icmp.Type) << 8) | uint16(seq.Icmp.Code))

		lays = append(lays, &icmph)
	}

	// Handle payload(s).
	pl := gopacket.Payload{}

	// Check if we have a static payload.
	if len(pl) == 1 {
		v := seq.Payloads[0]

		if len(v.Exact) > 0 {
			if v.IsString {
				pl = []byte(v.Exact)
			} else {
				data, err := utils.HexadecimalsToBytes(v.Exact)

				if err != nil {
					return fmt.Errorf("failed to parse static payload data as hexadecimal: %v", err)
				}

				pl = data
			}
		} else if v.MinLen == v.MaxLen && v.IsStatic {
			pl = utils.GenRandBytesSingle(int(v.MinLen), rng)
		}
	}

	// Handle packet counters.
	nextCounterUpdate := time.Now().Unix() + 1

	pps := seq.Pps
	curPps := uint64(0)

	bps := seq.Bps
	curBps := uint64(0)

	totPkts := uint64(0)
	totBytes := uint64(0)

	for k := range threads {
		cfg.DebugMsg(1, "Spawning thread #%d for sequence...", k)

		// Spawn thread.
		go func() {
			for {
				// Retrieve current time.
				now := time.Now().Unix()

				// Retrieve packet length.
				pktLen := 0

				// Check packet counters.
				if now > nextCounterUpdate {
					nextCounterUpdate = now + 1

					curPps = 1
					curBps = uint64(pktLen)
				} else {
					curPps++
					curBps += uint64(pktLen)
				}

				totPkts++
				totBytes += uint64(pktLen)

				if totPkts > seq.MaxPkts {
					return
				}

				if totBytes > seq.MaxBytes {
					return
				}
			}
		}()
	}

	// Cleanup tech.
	switch seq.Tech {
	case "af_packet":
		err := cAfpacket.Cleanup(int(seq.Threads))

		if err != nil {
			return fmt.Errorf("failed to cleanup AF_PACKET sequence: %v", err)
		}

	case "dpdk":
		err := cDpdk.Cleanup(int(seq.Threads))

		if err != nil {
			return fmt.Errorf("failed to cleanup DPDK sequence: %v", err)
		}

	default:
		err := cAfxdp.Cleanup(int(seq.Threads))

		if err != nil {
			return fmt.Errorf("failed to cleanup AF_XDP sequence: %v", err)
		}
	}

	return nil
}
