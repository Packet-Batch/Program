package sequence

import (
	"fmt"

	"github.com/Packet-Batch/Program/internal/tech/afxdp"
	"github.com/Packet-Batch/Program/internal/utils"
	"github.com/google/gopacket"
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
		ctx.CheckTime()

		// Regenerate seed if needed (every 10,000 nanoseconds to try to save CPU cycles).
		ctx.CheckRand()

		// Check packet rates.
		if ctx.CheckPktRates() {
			continue
		}

		// Check for random source IP from range.
		if ctx.CheckSrcIpRanges() {
			continue
		}

		// Generate random TTL and ID if needed.
		ctx.CheckTtl()
		ctx.CheckId()

		// Handle layer-4 protocols.
		ctx.CheckPorts()

		// Check if we need to regenerate payload.
		if ctx.CheckPl() {
			continue
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
		ctx.IncPktCounters(pktLen)

		ctx.Cfg.DebugMsg(5, "[SEQ %d] Send packet from '%s' to '%s' on thread #%d (length => %d, current PPS => %d, current BPS =>  %d)...", ctx.SeqIdx, ctx.Iph.SrcIP.String(), ctx.Iph.DstIP.String(), ctx.Id, pktLen, *ctx.CurPps, *ctx.CurBps)

		// Check total counters and limits.
		if ctx.CheckTotals() {
			break
		}

		// Alternate payload if there are mulitple.
		ctx.AlternatePl()

		utils.SleepMicro(ctx.Seq.Delay)
	}

	// Cleanup.
	err = aCtx.Cleanup(sock)

	if err != nil {
		return fmt.Errorf("failed to cleanup AF_XDP socket: %v", err)
	}

	return nil
}
