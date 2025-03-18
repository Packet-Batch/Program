package sequence

import (
	"fmt"
	"math/rand"

	"github.com/Packet-Batch/Program/internal/network"
	"github.com/Packet-Batch/Program/internal/tech/afxdp"
	"github.com/Packet-Batch/Program/internal/utils"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func (ctx *Sequence) SeqAfXdp() error {
	var err error

	// Create AF_XDP context.
	aCtx, err := afxdp.New()

	if err != nil {
		return fmt.Errorf("failed to setup AF_XDP context: %v", err)
	}

	// Create AF_XDP socket.
	queueId := ctx.Cli.AfXdp.Queue

	if queueId < 0 {
		queueId = int(ctx.TIdx)
	}

	sock, err := aCtx.Setup(ctx.Dev, queueId, ctx.Cli.AfXdp.NeedWakeup, ctx.Cli.AfXdp.SharedUmem, ctx.Cli.AfXdp.ForceSkb, ctx.Cli.AfXdp.ZeroCopy)

	if err != nil {
		return fmt.Errorf("faield to setup AF_XDP socket: %v", err)
	}

	// AF_XDP settings.
	batchSize := ctx.Cli.AfXdp.BatchSize

	for {
		// Retrieve current time if needed.
		if ctx.NeedNewTime {
			*ctx.Now, _ = utils.GetBootTimeNS()
		}

		// Regenerate seed if needed (every 10,000 nanoseconds to try to save CPU cycles).
		if ctx.NeedNewRand {
			if *ctx.Now > *ctx.NextRand {
				ctx.Rng = rand.New(rand.NewSource(*ctx.Now))

				*ctx.NextRand = *ctx.Now + ctx.RandInterval
			}
		}

		// Check packet rates.
		if ctx.DoPps || ctx.DoBps || ctx.Seq.Track {
			if *ctx.Now >= *ctx.NextCounterUpdate {
				*ctx.NextCounterUpdate = *ctx.Now + 1e9

				*ctx.CurPps = 0
				*ctx.CurBps = 0
			} else {
				// Check PPS and BPS rate limits.
				if ctx.DoPps && *ctx.CurPps >= ctx.Seq.Pps {
					utils.SleepMicro(ctx.Seq.Delay)

					continue
				}

				if ctx.DoBps && *ctx.CurBps >= ctx.Seq.Bps {
					utils.SleepMicro(ctx.Seq.Delay)

					continue
				}
			}
		}

		// Check for random source IP from range.
		if len(ctx.Seq.Ip4.SrcIpRanges) > 0 {
			ipRange := ""

			if len(ctx.Seq.Ip4.SrcIpRanges) == 1 {
				ipRange = ctx.Seq.Ip4.SrcIpRanges[0]
			} else {
				randIdx := ctx.Rng.Intn(len(ctx.Seq.Ip4.SrcIpRanges))

				ipRange = ctx.Seq.Ip4.SrcIpRanges[randIdx]
			}

			randIp, err := network.GetIpFromRange(ipRange, ctx.Rng)

			if err != nil {
				ctx.Cfg.DebugMsg(1, "[SEQ %d] Failed to retrieve random source IP from range '%s': %v", ctx.SeqIdx, ipRange, err)

				utils.SleepMicro(ctx.Seq.Delay)

				continue
			}

			ctx.Iph.SrcIP = network.U32ToNetIp(randIp)
		}

		// Generate random TTL and ID if needed.
		if ctx.Seq.Ip4.MinTtl != ctx.Seq.Ip4.MaxTtl {
			ctx.Iph.TTL = uint8(utils.GetRandInt(int(ctx.Seq.Ip4.MinTtl), int(ctx.Seq.Ip4.MaxTtl), ctx.Rng))
		}

		if ctx.Seq.Ip4.MinId != ctx.Seq.Ip4.MaxId {
			ctx.Iph.Id = uint16(utils.GetRandInt(int(ctx.Seq.Ip4.MinId), int(ctx.Seq.Ip4.MaxId), ctx.Rng))
		}

		// Handle layer-4 protocols.
		switch ctx.Iph.Protocol {
		case layers.IPProtocolTCP:
			// Generate random TCP ports if needed.
			if ctx.Seq.Tcp.SrcPort < 1 {
				ctx.Tcph.SrcPort = layers.TCPPort(uint16(utils.GetRandInt(1, 65535, ctx.Rng)))
			}

			if ctx.Seq.Tcp.DstPort < 1 {
				ctx.Tcph.DstPort = layers.TCPPort(uint16(utils.GetRandInt(1, 65535, ctx.Rng)))
			}

		case layers.IPProtocolUDP:
			// Generate random UDP ports if needed.
			if ctx.Seq.Udp.SrcPort < 1 {
				ctx.Udph.SrcPort = layers.UDPPort(uint16(utils.GetRandInt(1, 65535, ctx.Rng)))
			}

			if ctx.Seq.Udp.DstPort < 1 {
				ctx.Udph.DstPort = layers.UDPPort(uint16(utils.GetRandInt(1, 65535, ctx.Rng)))
			}
		}

		// Check if we need to regenerate payload.
		if ctx.CurPl != nil {
			if len(ctx.Seq.Payloads) > 1 || (len(ctx.CurPl.Exact) < 1 && !ctx.CurPl.IsStatic && ctx.CurPl.MaxLen > 0) {
				if len(ctx.CurPl.Exact) > 0 {
					if ctx.CurPl.IsFile {
						fData, err := utils.ReadFileAndStoreBytes(ctx.CurPl.Exact)

						if err != nil {
							ctx.Cfg.DebugMsg(1, "[SEQ %d] Failed to read payload data from file for payload #%d (file => %s): %v", ctx.SeqIdx, *ctx.CurPlIdx, ctx.CurPl.Exact, err)

							utils.SleepMicro(ctx.Seq.Delay)

							continue
						}

						if ctx.CurPl.IsString {
							*ctx.Pl = gopacket.Payload(fData)
						} else {
							data, err := utils.HexadecimalsToBytes(string(fData))

							if err != nil {
								ctx.Cfg.DebugMsg(1, "[SEQ %d] Failed to parse payload data from file '%s' in hexadecimal: %v", ctx.SeqIdx, ctx.CurPl.Exact, err)

								utils.SleepMicro(ctx.Seq.Delay)

								continue
							}

							*ctx.Pl = gopacket.Payload(data)
						}

					} else {
						if ctx.CurPl.IsString {
							*ctx.Pl = []byte(ctx.CurPl.Exact)
						} else {
							data, err := utils.HexadecimalsToBytes(ctx.CurPl.Exact)

							if err != nil {
								ctx.Cfg.DebugMsg(1, "[SEQ %d] Failed to parse payload data as hexadecimal: %v", ctx.SeqIdx, err)

								utils.SleepMicro(ctx.Seq.Delay)

								continue
							}

							*ctx.Pl = gopacket.Payload(data)
						}
					}
				} else {
					data := utils.GenRandBytes(int(ctx.CurPl.MinLen), int(ctx.CurPl.MaxLen), ctx.Rng)

					*ctx.Pl = gopacket.Payload(data)
				}

				// We need to replace old payload reference.
				(*ctx.PktLayers)[len(*ctx.PktLayers)-1] = ctx.Pl
			}
		}

		// Serialize data.
		gopacket.SerializeLayers(*ctx.Buf, *ctx.PktOpts, *ctx.PktLayers...)

		// Ethernet header can add trailing bytes that are zero-padded.
		// Make sure we're ignoring those.

		pkt := (*ctx.Buf).Bytes()
		pktLen := int(ctx.Iph.Length) + 14

		err = aCtx.SendPacket(sock, pkt, pktLen, batchSize)

		if err != nil {
			ctx.Cfg.DebugMsg(1, "[SEQ %d] Failed to send packet on thread #%d: %v", ctx.SeqIdx, ctx.Id, err)

			utils.SleepMicro(ctx.Seq.Delay)

			continue
		}

		// Increment packet counters.
		if ctx.DoPps || ctx.DoBps || ctx.Seq.Track {
			*ctx.CurPps++
			*ctx.CurBps += uint64(pktLen)

			*ctx.TotPkts++
			*ctx.TotBytes += uint64(pktLen)
		}

		ctx.Cfg.DebugMsg(5, "[SEQ %d] Send packet from '%s' to '%s' on thread #%d (length => %d, current PPS => %d, current BPS =>  %d)...", ctx.SeqIdx, ctx.Iph.SrcIP.String(), ctx.Iph.DstIP.String(), ctx.Id, pktLen, *ctx.CurPps, *ctx.CurBps)

		// Check total counters and limits.
		if ctx.Seq.MaxPkts > 0 && *ctx.TotPkts > ctx.Seq.MaxPkts {
			break
		}

		if ctx.Seq.MaxBytes > 0 && *ctx.TotBytes > ctx.Seq.MaxBytes {
			break
		}

		// Check time.
		if ctx.EndTime > 0 && *ctx.Now > ctx.EndTime {
			break
		}

		// Alternate payload if there are mulitple.
		if ctx.CurPl != nil && len(ctx.Seq.Payloads) > 1 {

			*ctx.CurPlIdx = (*ctx.CurPlIdx + 1) % len(ctx.Seq.Payloads)

			ctx.CurPl = &ctx.Seq.Payloads[*ctx.CurPlIdx]
		}

		utils.SleepMicro(ctx.Seq.Delay)
	}

	// Cleanup.
	err = aCtx.Cleanup(sock)

	if err != nil {
		return fmt.Errorf("failed to cleanup AF_XDP socket: %v", err)
	}

	return nil
}
