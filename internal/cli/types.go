package cli

type SeqOverride struct {
	Interface string

	SrcMac string
	DstMac string

	Protocol    string
	SrcIp       string
	SrcIpRanges string
	DstIp       string

	Tos int

	MinTtl int
	MaxTtl int

	MinId int
	MaxId int

	L3Csum int
	L4Csum int

	SrcPort int
	DstPort int

	TcpCooked  int
	TcpOneConn int

	TcpSyn int
	TcpAck int
	TcpPsh int
	TcpFin int
	TcpRst int
	TcpUrg int
	TcpEce int
	TcpCwr int

	IcmpCode int
	IcmpType int

	PlMinLen int
	PlMaxLen int

	PlStatic int
	PlFile   int
	PlString int
	PlExact  string
}

type AfXdp struct {
	Queue      int
	NeedWakeup bool
	SharedUmem bool
	BatchSize  int
	ForceSkb   bool
	ZeroCopy   bool
}

type Dpdk struct {
	LCores    string
	PortMask  string
	Queues    int
	Promisc   bool
	BurstSize int
}

type Cli struct {
	Cfg  string
	List bool
	Tech string

	SeqOverride SeqOverride

	AfXdp AfXdp
	Dpdk  Dpdk
}
