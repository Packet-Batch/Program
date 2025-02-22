package sequence

import (
	"fmt"
	"time"

	"github.com/Packet-Batch/Program/internal/cli"
	"github.com/Packet-Batch/Program/internal/config"
	"github.com/Packet-Batch/Program/internal/network"
	"github.com/Packet-Batch/Program/internal/tech"
	"github.com/Packet-Batch/Program/internal/tech/afpacket"
	"github.com/Packet-Batch/Program/internal/tech/afxdp"
	"github.com/Packet-Batch/Program/internal/tech/dpdk"
	"github.com/Packet-Batch/Program/internal/utils"
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

	// Get thread count.
	threads := seq.Threads

	if threads < 1 {
		threads = uint8(utils.GetCpuCount())
	}

	if threads < 1 {
		return fmt.Errorf("threads below 1")
	}

	// Retrieve source MAC address.
	srcMac := [6]byte{}

	if seq.Eth.SrcMac != nil {
		srcMac, err = network.MacAddrStrToArr(*seq.Eth.SrcMac)

		if err != nil {
			return fmt.Errorf("failed to parse source MAC address: %v", err)
		}
	} else {
		srcMac = network.GetMacOfInterface(*dev)
	}

	// Retrieve destination MAC address.
	dstMac := [6]byte{}

	if seq.Eth.DstMac != nil {
		dstMac, err = network.MacAddrStrToArr(*seq.Eth.DstMac)

		if err != nil {
			return fmt.Errorf("failed to parse destination MAC address: %v", err)
		}
	} else {
		dstMac = network.GetGatewayMacAddr()
	}

	cfg.DebugMsg(3, "[SEQ] Using src MAC => %x:%x:%x:%x:%x:%x", srcMac[0], srcMac[1], srcMac[2], srcMac[3], srcMac[4], srcMac[5])
	cfg.DebugMsg(3, "[SEQ] Using dst MAC => %x:%x:%x:%x:%x:%x", dstMac[0], dstMac[1], dstMac[2], dstMac[3], dstMac[4], dstMac[5])

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
		err := cAfxdp.Setup(*dev, cli.AfXdp.Queue, cli.AfXdp.NeedWakeup, cli.AfXdp.SharedUmem, cli.AfXdp.ForceSkb, cli.AfXdp.ZeroCopy, int(threads))

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
