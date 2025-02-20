package sequence

import (
	"fmt"
	"time"

	"github.com/Packet-Batch/Program/internal/cli"
	"github.com/Packet-Batch/Program/internal/config"
	"github.com/Packet-Batch/Program/internal/tech"
	"github.com/Packet-Batch/Program/internal/tech/afpacket"
	"github.com/Packet-Batch/Program/internal/tech/afxdp"
	"github.com/Packet-Batch/Program/internal/tech/dpdk"
)

func ProcessSeq(cfg *config.Config, cli *cli.Cli, seq *config.Sequence) error {
	// Load tech.
	t, err := tech.Load(seq.Tech)

	if err != nil {
		return fmt.Errorf("failed to load tech '%s': %v", seq.Tech, err)
	}

	dev := seq.Interface

	if dev == nil {
		dev = cfg.Interface

		if dev == nil {
			return fmt.Errorf("no interface found in sequence or config")
		}
	}

	cAfxdp, _ := t.(*afxdp.Context)
	cAfpacket, _ := t.(*afpacket.Context)
	cDpdk, _ := t.(*dpdk.Context)

	// Setup tech.
	switch seq.Tech {
	case "af_packet":
		err := cAfpacket.Setup(*dev, seq.Tcp.UseCookedSocket, int(seq.Threads))

		if err != nil {
			return fmt.Errorf("failed to setup AF_PACKET sequence: %v", err)
		}

	case "dpdk":
		err := cDpdk.Setup(*dev, int(seq.Threads))

		if err != nil {
			return fmt.Errorf("failed to setup DPDK sequence: %v", err)
		}
	default:
		err := cAfxdp.Setup(*dev, cli.AfXdp.Queue, cli.AfXdp.NeedWakeup, cli.AfXdp.SharedUmem, cli.AfXdp.ForceSkb, cli.AfXdp.ZeroCopy, int(seq.Threads))

		if err != nil {
			return fmt.Errorf("failed to setup AF_XDP sequence: %v", err)
		}
	}

	cnt := 0
	for {
		cnt++

		if cnt > 10 {
			break
		}

		data := []byte("HELLO WORLD")

		err := cAfxdp.SendPacket(data, len(data), 0, 1)

		if err != nil {
			cfg.DebugMsg(1, "[SEQ] Failed to send packet: %v", err)
		}

		time.Sleep(1 * time.Second)
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
			return fmt.Errorf("failed to cleanup DPDK sequence: %v")
		}

	default:
		err := cAfxdp.Cleanup(int(seq.Threads))

		if err != nil {
			return fmt.Errorf("failed to cleanup AF_XDP sequence: %v", err)
		}
	}

	return nil
}
