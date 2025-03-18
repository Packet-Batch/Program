package sequence

import (
	"math/rand"

	"github.com/Packet-Batch/Program/internal/cli"
	"github.com/Packet-Batch/Program/internal/config"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Sequence struct {
	Cli *cli.Cli
	Cfg *config.Config

	Dev string

	SeqIdx int
	TIdx   int
	Id     int

	NeedNewTime bool
	NeedNewRand bool

	Now *int64
	Rng *rand.Rand

	NextRand *int64

	RandInterval int64

	DoPps bool
	DoBps bool

	Seq   *config.Sequence
	CurPl *config.Payload

	CurPlIdx *int

	NextCounterUpdate *int64

	CurPps *uint64
	CurBps *uint64

	TotPkts  *uint64
	TotBytes *uint64

	EndTime int64

	Buf       *gopacket.SerializeBuffer
	PktOpts   *gopacket.SerializeOptions
	PktLayers *[]gopacket.SerializableLayer

	Eth   *layers.Ethernet
	Iph   *layers.IPv4
	Tcph  *layers.TCP
	Udph  *layers.UDP
	Icmph *layers.ICMPv4

	Pl *gopacket.Payload
}
