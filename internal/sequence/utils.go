package sequence

import (
	"math/rand"

	"github.com/Packet-Batch/Program/internal/network"
	"github.com/Packet-Batch/Program/internal/utils"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func (ctx *Sequence) CheckTime() {
	if ctx.NeedNewTime {
		*ctx.Now, _ = utils.GetBootTimeNS()
	}
}

func (ctx *Sequence) CheckRand() {
	if ctx.NeedNewRand {
		if *ctx.Now > *ctx.NextRand {
			ctx.Rng = rand.New(rand.NewSource(*ctx.Now))

			*ctx.NextRand = *ctx.Now + ctx.RandInterval
		}
	}
}

func (ctx *Sequence) CheckPktRates() bool {
	if ctx.DoPps || ctx.DoBps || ctx.Seq.Track {
		if *ctx.Now >= *ctx.NextCounterUpdate {
			*ctx.NextCounterUpdate = *ctx.Now + 1e9

			*ctx.CurPps = 0
			*ctx.CurBps = 0
		} else {
			// Check PPS and BPS rate limits.
			if ctx.DoPps && *ctx.CurPps >= ctx.Seq.Pps {
				utils.SleepMicro(ctx.Seq.Delay)

				return true
			}

			if ctx.DoBps && *ctx.CurBps >= ctx.Seq.Bps {
				utils.SleepMicro(ctx.Seq.Delay)

				return true
			}
		}
	}

	return false
}

func (ctx *Sequence) CheckSrcIpRanges() bool {
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

			return true
		}

		ctx.Iph.SrcIP = network.U32ToNetIp(randIp)
	}

	return false
}

func (ctx *Sequence) CheckTtl() {
	if ctx.Seq.Ip4.MinTtl != ctx.Seq.Ip4.MaxTtl {
		ctx.Iph.TTL = uint8(utils.GetRandInt(int(ctx.Seq.Ip4.MinTtl), int(ctx.Seq.Ip4.MaxTtl), ctx.Rng))
	}
}

func (ctx *Sequence) CheckId() {
	if ctx.Seq.Ip4.MinId != ctx.Seq.Ip4.MaxId {
		ctx.Iph.Id = uint16(utils.GetRandInt(int(ctx.Seq.Ip4.MinId), int(ctx.Seq.Ip4.MaxId), ctx.Rng))
	}
}

func (ctx *Sequence) CheckPorts() {
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
}

func (ctx *Sequence) CheckPl() bool {
	if ctx.CurPl != nil {
		if len(ctx.Seq.Payloads) > 1 || (len(ctx.CurPl.Exact) < 1 && !ctx.CurPl.IsStatic && ctx.CurPl.MaxLen > 0) {
			if len(ctx.CurPl.Exact) > 0 {
				if ctx.CurPl.IsFile {
					fData, err := utils.ReadFileAndStoreBytes(ctx.CurPl.Exact)

					if err != nil {
						ctx.Cfg.DebugMsg(1, "[SEQ %d] Failed to read payload data from file for payload #%d (file => %s): %v", ctx.SeqIdx, *ctx.CurPlIdx, ctx.CurPl.Exact, err)

						utils.SleepMicro(ctx.Seq.Delay)

						return true
					}

					if ctx.CurPl.IsString {
						*ctx.Pl = gopacket.Payload(fData)
					} else {
						data, err := utils.HexadecimalsToBytes(string(fData))

						if err != nil {
							ctx.Cfg.DebugMsg(1, "[SEQ %d] Failed to parse payload data from file '%s' in hexadecimal: %v", ctx.SeqIdx, ctx.CurPl.Exact, err)

							utils.SleepMicro(ctx.Seq.Delay)

							return true
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

							return true
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

	return false
}

func (ctx *Sequence) IncPktCounters(pktLen int) {
	if ctx.DoPps || ctx.DoBps || ctx.Seq.Track {
		*ctx.CurPps++
		*ctx.CurBps += uint64(pktLen)

		*ctx.TotPkts++
		*ctx.TotBytes += uint64(pktLen)
	}
}

func (ctx *Sequence) CheckTotals() bool {
	// Check total counters and limits.
	if ctx.Seq.MaxPkts > 0 && *ctx.TotPkts > ctx.Seq.MaxPkts {
		return true
	}

	if ctx.Seq.MaxBytes > 0 && *ctx.TotBytes > ctx.Seq.MaxBytes {
		return true
	}

	// Check time.
	if ctx.EndTime > 0 && *ctx.Now > ctx.EndTime {
		return true
	}

	return false
}

func (ctx *Sequence) AlternatePl() {
	if ctx.CurPl != nil && len(ctx.Seq.Payloads) > 1 {
		*ctx.CurPlIdx = (*ctx.CurPlIdx + 1) % len(ctx.Seq.Payloads)

		ctx.CurPl = &ctx.Seq.Payloads[*ctx.CurPlIdx]
	}
}
