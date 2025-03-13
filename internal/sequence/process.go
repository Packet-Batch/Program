package sequence

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"math/rand"

	"github.com/Packet-Batch/Program/internal/cli"
	"github.com/Packet-Batch/Program/internal/config"
	"github.com/Packet-Batch/Program/internal/network"
	"github.com/Packet-Batch/Program/internal/tech"
	"github.com/Packet-Batch/Program/internal/tech/afxdp"
	"github.com/Packet-Batch/Program/internal/utils"
	"github.com/google/gopacket"

	"github.com/google/gopacket/layers"
)

func ProcessSeq(cfg *config.Config, cli *cli.Cli, seq *config.Sequence, idx int) error {
	var err error

	// Generate base random seed.
	rngBase := rand.New(rand.NewSource(time.Now().UnixNano()))

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
	nextCounterUpdate := time.Now().Unix() + 1

	curPps := uint64(0)
	curBps := uint64(0)

	totPkts := uint64(0)
	totBytes := uint64(0)

	// Get end time if needed.
	endTime := int64(0)

	if seq.Time > 0 {
		endTime = time.Now().Unix() + int64(seq.Time)
	}

	var wg sync.WaitGroup

	cAfxdp, _ := t.(*afxdp.Context)

	for k := range threads {
		cfg.DebugMsg(1, "[SEQ %d] Spawning thread #%d for sequence...", idx, k)

		var curPl *config.Payload = nil
		curPlIdx := 0

		if len(seq.Payloads) > 0 {
			curPl = &seq.Payloads[0]
		}

		wg.Add(1)

		// Spawn thread.
		go func(k uint8) {
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

			// Create socket.
			queueId := cli.AfXdp.Queue

			if queueId < 0 {
				queueId = int(k)
			}

			sock, err := cAfxdp.Setup(dev, queueId, cli.AfXdp.NeedWakeup, cli.AfXdp.SharedUmem, cli.AfXdp.ForceSkb, cli.AfXdp.ZeroCopy)

			if err != nil {
				cfg.DebugMsg(1, "[SEQ %d] Failed to create AF_XDP socket on thread #%d: %v", seqIdx, id, err)

				return
			}

			// AF_XDP settings.
			batchSize := cli.AfXdp.BatchSize

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

			do_pps := false

			if seqLoc.Pps > 0 {
				do_pps = true
			}

			do_bps := false

			if seqLoc.Bps > 0 {
				do_bps = true
			}

			for {
				// Regenerate seed.
				rng := rand.New(rand.NewSource(time.Now().UnixNano()))

				// Retrieve current time.
				now := time.Now().Unix()

				// Check packet rates.
				if do_pps || do_bps || seqLoc.Track {
					if now >= nextCounterUpdate {
						nextCounterUpdate = now + 1

						curPps = 0
						curBps = 0
					} else {
						// Check PPS and BPS rate limits.
						if do_pps && curPps >= seqLoc.Pps {
							utils.SleepMicro(seqLoc.Delay)

							continue
						}

						if do_bps && curBps >= seqLoc.Bps {
							utils.SleepMicro(seqLoc.Delay)

							continue
						}
					}
				}

				// Check for random source IP from range.
				if len(seqLoc.Ip4.SrcIpRanges) > 0 {
					ipRange := ""

					if len(seqLoc.Ip4.SrcIpRanges) == 1 {
						ipRange = seqLoc.Ip4.SrcIpRanges[0]
					} else {
						randIdx := rng.Intn(len(seqLoc.Ip4.SrcIpRanges))

						ipRange = seqLoc.Ip4.SrcIpRanges[randIdx]
					}

					randIp, err := network.GetIpFromRange(ipRange, rng)

					if err != nil {
						cfgLoc.DebugMsg(1, "[SEQ %d] Failed to retrieve random source IP from range '%s': %v", seqIdx, ipRange, err)

						utils.SleepMicro(seqLoc.Delay)

						continue
					}

					iph.SrcIP = network.U32ToNetIp(randIp)
				}

				// Generate random TTL and ID if needed.
				if seqLoc.Ip4.MinTtl != seqLoc.Ip4.MaxTtl {
					iph.TTL = uint8(utils.GetRandInt(int(seqLoc.Ip4.MinTtl), int(seqLoc.Ip4.MaxTtl), rng))
				}

				if seqLoc.Ip4.MinId != seqLoc.Ip4.MaxId {
					iph.Id = uint16(utils.GetRandInt(int(seqLoc.Ip4.MinId), int(seqLoc.Ip4.MaxId), rng))
				}

				// Handle layer-4 protocols.
				switch iph.Protocol {
				case layers.IPProtocolTCP:
					// Generate random TCP ports if needed.
					if seqLoc.Tcp.SrcPort < 1 {
						tcph.SrcPort = layers.TCPPort(uint16(utils.GetRandInt(1, 65535, rng)))
					}

					if seqLoc.Tcp.DstPort < 1 {
						tcph.DstPort = layers.TCPPort(uint16(utils.GetRandInt(1, 65535, rng)))
					}

				case layers.IPProtocolUDP:
					// Generate random UDP ports if needed.
					if seqLoc.Udp.SrcPort < 1 {
						udph.SrcPort = layers.UDPPort(uint16(utils.GetRandInt(1, 65535, rng)))
					}

					if seqLoc.Udp.DstPort < 1 {
						udph.DstPort = layers.UDPPort(uint16(utils.GetRandInt(1, 65535, rng)))
					}
				}

				// Check if we need to regenerate payload.
				if curPl != nil {
					if len(seqLoc.Payloads) > 1 || (len(curPl.Exact) < 1 && curPl.MinLen != curPl.MaxLen && !curPl.IsStatic) {
						if len(curPl.Exact) > 0 {
							if curPl.IsFile {
								fData, err := utils.ReadFileAndStoreBytes(curPl.Exact)

								if err != nil {
									cfgLoc.DebugMsg(1, "[SEQ %d] Failed to read payload data from file for payload #%d (file => %s): %v", seqIdx, curPlIdx, curPl.Exact, err)

									utils.SleepMicro(seqLoc.Delay)

									continue
								}

								if curPl.IsString {
									pl = gopacket.Payload(fData)
								} else {
									data, err := utils.HexadecimalsToBytes(string(fData))

									if err != nil {
										cfgLoc.DebugMsg(1, "[SEQ %d] Failed to parse payload data from file '%s' in hexadecimal: %v", seqIdx, curPl.Exact, err)

										utils.SleepMicro(seqLoc.Delay)

										continue
									}

									pl = gopacket.Payload(data)
								}

							} else {
								if curPl.IsString {
									pl = []byte(curPl.Exact)
								} else {
									data, err := utils.HexadecimalsToBytes(curPl.Exact)

									if err != nil {
										cfgLoc.DebugMsg(1, "[SEQ %d] Failed to parse payload data as hexadecimal: %v", seqIdx, err)

										utils.SleepMicro(seqLoc.Delay)

										continue
									}

									pl = gopacket.Payload(data)
								}
							}
						} else {
							data := utils.GenRandBytes(int(curPl.MinLen), int(curPl.MaxLen), rng)

							pl = gopacket.Payload(data)
						}

						// We need to replace old payload reference.
						pktLayers[len(pktLayers)-1] = &pl
					}
				}

				// Serialize data.
				gopacket.SerializeLayers(buf, pktOpts, pktLayers...)

				// Ethernet header can add trailing bytes that are zero-padded.
				// Make sure we're ignoring those.

				pkt := buf.Bytes()
				pktLen := int(iph.Length) + 14

				err = cAfxdp.SendPacket(sock, pkt, pktLen, batchSize)

				if err != nil {
					cfgLoc.DebugMsg(1, "[SEQ %d] Failed to send packet on thread #%d: %v", seqIdx, k+1, err)

					utils.SleepMicro(seqLoc.Delay)

					continue
				}

				// Increment packet counters.
				if do_pps || do_bps || seqLoc.Track {
					curPps++
					curBps += uint64(pktLen)

					totPkts++
					totBytes += uint64(pktLen)
				}

				cfgLoc.DebugMsg(5, "[SEQ %d] Send packet from '%s' to '%s' on thread #%d (length => %d, current PPS => %d, current BPS =>  %d)...", seqIdx, iph.SrcIP.String(), iph.DstIP.String(), id, pktLen, curPps, curBps)

				// Check total counters and limits.
				if seqLoc.MaxPkts > 0 && totPkts > seqLoc.MaxPkts {
					break
				}

				if seqLoc.MaxBytes > 0 && totBytes > seqLoc.MaxBytes {
					break
				}

				// Check time.
				if endTime > 0 && now > endTime {
					break
				}

				// Alternate payload if there are mulitple.
				if curPl != nil && len(seqLoc.Payloads) > 1 {

					curPlIdx = (curPlIdx + 1) % len(seqLoc.Payloads)

					curPl = &seqLoc.Payloads[curPlIdx]
				}

				utils.SleepMicro(seqLoc.Delay)
			}

			// Cleanup socket.
			err = cAfxdp.Cleanup(sock)

			if err != nil {
				cfgLoc.DebugMsg(1, "[SEQ %d] Failed to cleanup AF_XDP socket on thread #%d: %v", seqIdx, id, err)
			}
		}(k)
	}

	if seq.Block {
		wg.Wait()
	}

	return nil
}
