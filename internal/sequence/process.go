package sequence

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"math/rand"

	"github.com/Packet-Batch/Program/internal/cli"
	"github.com/Packet-Batch/Program/internal/config"
	"github.com/Packet-Batch/Program/internal/network"
	"github.com/Packet-Batch/Program/internal/utils"
	"github.com/google/gopacket"

	"github.com/google/gopacket/layers"
)

func ProcessSeq(cfg *config.Config, cli *cli.Cli, seq *config.Sequence, idx int) error {
	var err error

	// Determine tech to use.
	tech := cli.Tech

	if len(tech) < 1 {
		tech = seq.Tech

		if len(tech) < 1 {
			return fmt.Errorf("no tech set in sequence or CLI override")
		}
	}

	// Determine interface to use.
	dev := seq.Interface

	if len(dev) < 1 {
		dev = cfg.Interface

		if len(dev) < 1 {
			return fmt.Errorf("no interface set in sequence or config")
		}
	}

	// Determine thread count.
	threads := seq.Threads

	if threads < 1 {
		threads = utils.GetCpuCount()
	}

	if threads < 1 {
		return fmt.Errorf("thread count below 1 somehow (this is not normal)")
	}

	// Retrieve source MAC address.
	var srcMac []byte

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
	var dstMac []byte

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

	cfg.DebugMsg(3, "[SEQ %d] Using src MAC => %x:%x:%x:%x:%x:%x", idx, srcMac[0], srcMac[1], srcMac[2], srcMac[3], srcMac[4], srcMac[5])
	cfg.DebugMsg(3, "[SEQ %d] Using dst MAC => %x:%x:%x:%x:%x:%x", idx, dstMac[0], dstMac[1], dstMac[2], dstMac[3], dstMac[4], dstMac[5])

	// Determine static source IP if set.
	var srcIp *net.IPAddr = nil

	if len(seq.Ip4.SrcIp) > 0 {
		srcIp, err = net.ResolveIPAddr("ip", seq.Ip4.SrcIp)

		if err != nil {
			return fmt.Errorf("failed to resolve source Ipv4 address '%s': %v", seq.Ip4.SrcIp, err)
		}
	}

	if srcIp == nil && len(seq.Ip4.SrcIpRanges) < 1 {
		return fmt.Errorf("no source IP or ranges specified")
	}

	// Determine destination IP.
	if len(seq.Ip4.DstIp) < 1 {
		return fmt.Errorf("no destination IPv4 address specified")
	}

	dstIp, err := net.ResolveIPAddr("ip", seq.Ip4.DstIp)

	if err != nil {
		return fmt.Errorf("failed to resolve destination IP address'%s': %v", seq.Ip4.DstIp, err)
	}

	// Check and set IP protocol.
	proto, err := network.GetProtocolIdByStr(seq.Ip4.Protocol)

	if err != nil {
		return fmt.Errorf("failed to find protocol by string: %v", err)
	}

	// Generate base random seed.
	rngBase := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Determine static payload.
	var staticPl []byte

	if len(seq.Payloads) > 0 {
		if len(seq.Payloads) == 1 {
			v := seq.Payloads[0]

			if len(v.Exact) > 0 {
				if v.IsString {
					staticPl = []byte(v.Exact)
				} else {
					data, err := utils.HexadecimalsToBytes(v.Exact)

					if err != nil {
						return fmt.Errorf("failed to parse static payload data as hexadecimal: %v", err)
					}

					staticPl = data
				}
			} else if v.MinLen == v.MaxLen && v.IsStatic {
				staticPl = utils.GenRandBytesSingle(int(v.MinLen), rngBase)
			}
		}
	}

	// Handle packet counters.
	nextCounterUpdate, _ := utils.GetBootTimeNS()

	nextCounterUpdate += 1e9

	curPps := uint64(0)
	curBps := uint64(0)

	totPkts := uint64(0)
	totBytes := uint64(0)

	// Get end time if needed.
	endTime := int64(0)

	if seq.Time > 0 {
		endTime, _ = utils.GetBootTimeNS()

		endTime += (int64(seq.Time) * 1e9)
	}

	var wg sync.WaitGroup

	for k := range threads {
		cfg.DebugMsg(1, "[SEQ %d] Spawning thread #%d for sequence...", idx, k)

		var curPl *config.Payload = nil
		curPlIdx := 0

		if len(seq.Payloads) > 0 {
			curPl = &seq.Payloads[0]
		}

		wg.Add(1)

		// Spawn thread.
		go func(k int) {
			defer wg.Done()

			var err error

			// Copy settings to new variables to avoid accessing shared memory between threads which would hurt performance.
			// Note - Not sure if Golang handles this automatically?
			// Either way, it's cleaner due to shorter variable names I guess...
			id := k + 1
			seqIdx := idx

			cfgLoc := &config.Config{}

			cfgData, err := json.Marshal(cfg)

			if err != nil {
				cfg.DebugMsg(1, "[SEQ %d] Failed to marshal config for local deep copy in thread #%d: %v", seqIdx, id, err)

				return
			}

			err = json.Unmarshal(cfgData, cfgLoc)

			if err != nil {
				cfg.DebugMsg(1, "[SEQ %d] Failed to unmarshal config for local deep copy in thread #%d: %v", seqIdx, id, err)

				return
			}

			if seqIdx-1 >= len(cfgLoc.Sequences) {
				cfgLoc.DebugMsg(1, "[SEQ %d] Sequence from config's deep copy at index %d is invalid.", seqIdx, seqIdx-1)

				return
			}

			seqLoc := &cfgLoc.Sequences[seqIdx-1]

			// Create packet buffer and options.
			buf := gopacket.NewSerializeBuffer()
			pktOpts := gopacket.SerializeOptions{
				ComputeChecksums: seqLoc.ComputeCsums,
				FixLengths:       true,
			}

			// Create ethernet layer.
			eth := &layers.Ethernet{
				EthernetType: layers.EthernetTypeIPv4,
				SrcMAC:       srcMac,
				DstMAC:       dstMac,
			}

			// Create IPv4 layer.
			iph := &layers.IPv4{
				Version:    4,
				IHL:        5,
				FragOffset: 0,
				TOS:        seqLoc.Ip4.Tos,
				DstIP:      dstIp.IP,
				Protocol:   proto,
				Length:     20,
			}

			// Create global layers and add ethernet and IP header to it.
			pktLayers := []gopacket.SerializableLayer{eth, iph}

			// Check for static IPv4 fields.
			if seqLoc.Ip4.MinTtl == seqLoc.Ip4.MaxTtl {
				iph.TTL = seqLoc.Ip4.MinTtl
			}

			if seqLoc.Ip4.MinId == seqLoc.Ip4.MaxId {
				iph.Id = seqLoc.Ip4.MinId
			}

			if srcIp != nil {
				iph.SrcIP = srcIp.IP
			}

			// Create and handle layer-4 layers.
			tcph := &layers.TCP{}
			udph := &layers.UDP{}
			icmph := &layers.ICMPv4{}

			switch proto {
			case layers.IPProtocolTCP:
				if seqLoc.Tcp.SrcPort > 0 {
					tcph.SrcPort = layers.TCPPort(seqLoc.Tcp.SrcPort)
				}

				if seqLoc.Tcp.DstPort > 0 {
					tcph.DstPort = layers.TCPPort(seqLoc.Tcp.DstPort)
				}

				tcph.SetNetworkLayerForChecksum(iph)

				pktLayers = append(pktLayers, tcph)

			case layers.IPProtocolUDP:
				if seqLoc.Udp.SrcPort > 0 {
					udph.SrcPort = layers.UDPPort(seqLoc.Udp.SrcPort)
				}

				if seqLoc.Udp.DstPort > 0 {
					udph.DstPort = layers.UDPPort(seqLoc.Udp.DstPort)
				}

				udph.SetNetworkLayerForChecksum(iph)

				pktLayers = append(pktLayers, udph)

			case layers.IPProtocolICMPv4:
				icmph.TypeCode = layers.ICMPv4TypeCode((uint16(seqLoc.Icmp.Type) << 8) | uint16(seqLoc.Icmp.Code))

				pktLayers = append(pktLayers, icmph)
			}

			// Create payload.
			var pl gopacket.Payload

			// Check for static payload.
			if len(staticPl) > 0 {
				// Avoid shared memory access by deep copying before loop.
				staticPlCopy := make([]byte, len(staticPl))
				copy(staticPlCopy, staticPl)

				pl = gopacket.Payload(staticPlCopy)
			}

			pktLayers = append(pktLayers, &pl)

			doPps := false

			if seqLoc.Pps > 0 {
				doPps = true
			}

			doBps := false

			if seqLoc.Bps > 0 {
				doBps = true
			}

			// Retrieve time since boot.
			now, err := utils.GetBootTimeNS()

			if err != nil {
				cfgLoc.DebugMsg(1, "[SEQ %d] Failed to retrieve boot time in nanoseconds on thread #%d: %v", seqIdx, id, err)
			}

			needNewTime := false

			if doPps || doBps || seqLoc.Track || endTime > 0 {
				needNewTime = true
			}

			// Retrieve random seed interval.
			randInterval := seqLoc.RandInterval

			if randInterval < 0 {
				randInterval = 10000
			}

			// Handle random seeding.
			rng := rand.New(rand.NewSource(now))

			nextRand := now + randInterval

			needNewRand := false

			// Check for common settings that requires random seeding.
			if len(seqLoc.Ip4.SrcIpRanges) > 0 || seqLoc.Ip4.MinId != seqLoc.Ip4.MaxId || seqLoc.Ip4.MinTtl != seqLoc.Ip4.MaxTtl || (iph.Protocol == layers.IPProtocolTCP && (seqLoc.Tcp.SrcPort < 1 || seqLoc.Tcp.DstPort < 1)) || (iph.Protocol == layers.IPProtocolUDP && (seqLoc.Udp.SrcPort < 1 || seqLoc.Udp.DstPort < 1)) {
				needNewRand = true
			}

			// Check for payloads that need random seeding.
			if !needNewRand && len(seqLoc.Payloads) > 0 {
				for _, p := range seqLoc.Payloads {
					if len(p.Exact) < 1 && (!p.IsStatic && p.MaxLen > 0) {
						needNewRand = true

						break
					}
				}
			}

			// Make sure we retrieve new time if random seeding is enabled.
			if needNewRand && !needNewTime {
				needNewTime = true
			}

			// Create sequence context.
			ctx := &Sequence{
				Cli:         cli,
				Cfg:         cfgLoc,
				Dev:         dev,
				SeqIdx:      seqIdx,
				TIdx:        k,
				Id:          id,
				NeedNewTime: needNewTime,
				NeedNewRand: needNewRand,

				Now: &now,
				Rng: rng,

				NextRand:     &nextRand,
				RandInterval: randInterval,

				DoPps: doPps,
				DoBps: doBps,

				Seq:   seqLoc,
				CurPl: curPl,

				CurPlIdx: &curPlIdx,

				NextCounterUpdate: &nextCounterUpdate,

				CurPps: &curPps,
				CurBps: &curBps,

				TotPkts:  &totPkts,
				TotBytes: &totBytes,

				EndTime: endTime,

				Buf:       &buf,
				PktOpts:   &pktOpts,
				PktLayers: &pktLayers,

				Eth:   eth,
				Iph:   iph,
				Tcph:  tcph,
				Udph:  udph,
				Icmph: icmph,

				Pl: &pl,
			}

			// Determine which tech to use and then execute + initialize.
			techL := strings.ToLower(tech)

			switch techL {
			case "af_xdp":
				err = ctx.SeqAfXdp()
			}

			if err != nil {
				cfgLoc.DebugMsg(0, "[SEQ %d] Failed to setup and use tech: %v", seqIdx, err)
			}
		}(k)
	}

	if seq.Block {
		wg.Wait()
	}

	return nil
}
