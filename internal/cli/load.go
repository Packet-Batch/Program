package cli

import "flag"

func (cli *Cli) Parse() {
	// General flags
	StringOpt(&cli.Cfg, "c", "cfg", "./conf.json", "Path to config file.")
	BoolOpt(&cli.List, "l", "list", false, "Prints config settings and exits.")
	StringOpt(&cli.Tech, "t", "tech", "", "The packet processing technology to use.")

	// Sequence override
	StringOpt(&cli.SeqOverride.Interface, "i", "interface", "", "The interface name")

	IntOpt(&cli.SeqOverride.Block, "b", "block", -1, "Whether to block or not.")

	Int64Opt(&cli.SeqOverride.MaxPkts, "", "max-pkts", -1, "The maximum packets to send.")
	Int64Opt(&cli.SeqOverride.MaxBytes, "", "max-bytes", -1, "The maximum bytes to send.")

	Int64Opt(&cli.SeqOverride.Pps, "", "pps", -1, "The maximum packets per second to send concurrently.")
	Int64Opt(&cli.SeqOverride.Bps, "", "bps", -1, "The maximum bytes per second to send concurrently.")

	IntOpt(&cli.SeqOverride.Time, "", "time", -1, "The maximum amount of time in seconds to run this sequence for.")
	Int64Opt(&cli.SeqOverride.Delay, "", "delay", -1, "The delay between sending packets on each thread in micro-seconds.")
	IntOpt(&cli.SeqOverride.Threads, "", "threads", -1, "The amount of threads to create that sends packets.")

	StringOpt(&cli.SeqOverride.SrcMac, "", "smac", "", "The source MAC address.")
	StringOpt(&cli.SeqOverride.DstMac, "", "dmac", "", "The destination MAC address.")

	StringOpt(&cli.SeqOverride.Protocol, "p", "protocol", "", "The layer-4 protocol.")
	StringOpt(&cli.SeqOverride.SrcIp, "s", "src", "", "The source IP address.")
	StringOpt(&cli.SeqOverride.SrcIpRanges, "r", "ranges", "", "The source IP ranges.")
	StringOpt(&cli.SeqOverride.DstIp, "d", "dst", "", "The destination IP address.")

	IntOpt(&cli.SeqOverride.Tos, "", "tos", -1, "The Type of Service.")
	IntOpt(&cli.SeqOverride.MinTtl, "", "minttl", -1, "The minimum TTL.")
	IntOpt(&cli.SeqOverride.MaxTtl, "", "maxttl", -1, "The maximum TTL.")
	IntOpt(&cli.SeqOverride.MinId, "", "minid", -1, "The minimum ID.")
	IntOpt(&cli.SeqOverride.MaxId, "", "maxid", -1, "The maximum ID.")

	IntOpt(&cli.SeqOverride.SrcPort, "", "sport", -1, "The layer-4 source port.")
	IntOpt(&cli.SeqOverride.DstPort, "", "dport", -1, "The layer-4 destination port.")

	IntOpt(&cli.SeqOverride.TcpCooked, "", "cooked", -1, "Use TCP cooked socket.")
	IntOpt(&cli.SeqOverride.TcpOneConn, "", "oneconn", -1, "Use one TCP connection.")

	IntOpt(&cli.SeqOverride.TcpSyn, "", "syn", -1, "Set the TCP SYN flag.")
	IntOpt(&cli.SeqOverride.TcpAck, "", "ack", -1, "Set the TCP ACK flag.")
	IntOpt(&cli.SeqOverride.TcpPsh, "", "psh", -1, "Set the TCP PSH flag.")
	IntOpt(&cli.SeqOverride.TcpFin, "", "fin", -1, "Set the TCP FIN flag.")
	IntOpt(&cli.SeqOverride.TcpRst, "", "rst", -1, "Set the TCP RST flag.")
	IntOpt(&cli.SeqOverride.TcpUrg, "", "urg", -1, "Set the TCP URG flag.")
	IntOpt(&cli.SeqOverride.TcpEce, "", "ece", -1, "Set the TCP ECE flag.")
	IntOpt(&cli.SeqOverride.TcpCwr, "", "cwr", -1, "Set the TCP CWR flag.")

	IntOpt(&cli.SeqOverride.IcmpCode, "", "code", -1, "The ICMP code.")
	IntOpt(&cli.SeqOverride.IcmpType, "", "type", -1, "The ICMP type.")

	IntOpt(&cli.SeqOverride.PlMinLen, "", "plmin", -1, "The minimum payload length.")
	IntOpt(&cli.SeqOverride.PlMaxLen, "", "plmax", -1, "The maximum payload length.")

	IntOpt(&cli.SeqOverride.PlStatic, "", "static", -1, "Whether payload is static.")
	IntOpt(&cli.SeqOverride.PlFile, "f", "file", -1, "Whether payload is file.")
	IntOpt(&cli.SeqOverride.PlString, "", "string", -1, "Whether payload is string.")
	StringOpt(&cli.SeqOverride.PlExact, "e", "exact", "", "The payload data.")

	// AF_XDP tech
	IntOpt(&cli.AfXdp.Queue, "", "queue", -1, "If set, will bind all AF_XDP sockets to this queue ID.")
	BoolOpt(&cli.AfXdp.NeedWakeup, "", "needwakeup", true, "If set, will use the no wakeup flag on AF_XDP sockets.")
	BoolOpt(&cli.AfXdp.SharedUmem, "", "shared", false, "If set, will use shared umem with AF_XDP sockets.")
	IntOpt(&cli.AfXdp.BatchSize, "", "bathsize", 32, "The AF_XDP batch size.")
	BoolOpt(&cli.AfXdp.ForceSkb, "", "skb", false, "If set, will force AF_XDP sockets to use SKB mode.")
	BoolOpt(&cli.AfXdp.ZeroCopy, "", "zerocopy", false, "If set, will use zero-copy mode with AF_XDP sockets.")

	// DPDK tech

	flag.Parse()
}
