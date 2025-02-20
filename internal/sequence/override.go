package sequence

import (
	"fmt"
	"strings"

	"github.com/Packet-Batch/Program/internal/cli"
	"github.com/Packet-Batch/Program/internal/config"
	"github.com/Packet-Batch/Program/internal/utils"
)

func CheckFirstSeqOverride(cfg *config.Config, cli *cli.Cli) bool {
	changed := false

	var seq *config.Sequence = nil

	if len(cfg.Sequences) > 0 {
		seq = &cfg.Sequences[0]
	} else {
		seq = &config.Sequence{}
	}

	or := &cli.SeqOverride

	if len(or.Interface) > 0 {
		changed = true

		seq.Interface = &or.Interface
	}

	if len(or.SrcMac) > 0 {
		changed = true

		seq.Eth.SrcMac = &or.SrcMac
	}

	if len(or.DstMac) > 0 {
		changed = true

		seq.Eth.DstMac = &or.DstMac
	}

	if len(or.Protocol) > 0 {
		changed = true

		seq.Ip4.Protocol = or.Protocol
	}

	if len(or.SrcIp) > 0 {
		changed = true

		seq.Ip4.SrcIp = &or.SrcIp
	}

	if len(or.SrcIpRanges) > 0 {
		changed = true

		rangeParts := strings.Split(or.SrcIpRanges, ",")

		for _, rPart := range rangeParts {
			ipParts := strings.Split(rPart, "/")

			ipStr := strings.TrimSpace(ipParts[0])
			cidrStr := "32"

			if len(ipParts) > 1 {
				cidrStr = strings.TrimSpace(ipParts[1])
			}

			seq.Ip4.SrcIpRanges = append(seq.Ip4.SrcIpRanges, fmt.Sprintf("%s/%s", ipStr, cidrStr))
		}
	}

	if len(or.DstIp) > 0 {
		changed = true

		seq.Ip4.DstIp = or.DstIp
	}

	if or.Tos > -1 {
		changed = true

		seq.Ip4.Tos = uint8(or.Tos)
	}

	if or.MinTtl > -1 {
		changed = true

		seq.Ip4.MinTtl = uint16(or.MinTtl)
	}

	if or.MaxTtl > -1 {
		changed = true

		seq.Ip4.MaxTtl = uint16(or.MaxTtl)
	}

	if or.MinId > -1 {
		changed = true

		seq.Ip4.MinId = uint16(or.MinId)
	}

	if or.MaxId > -1 {
		changed = true

		seq.Ip4.MaxId = uint16(or.MaxId)
	}

	if or.L3Csum > -1 {
		changed = true

		seq.Ip4.Csum = utils.IntToBool(or.L3Csum)
	}

	if or.L4Csum > -1 {
		changed = true

		seq.Tcp.Csum = utils.IntToBool(or.L4Csum)
		seq.Udp.Csum = utils.IntToBool(or.L4Csum)
		seq.Icmp.Csum = utils.IntToBool(or.L4Csum)
	}

	if or.SrcPort > -1 {
		changed = true

		seq.Tcp.SrcPort = uint16(or.SrcPort)
		seq.Udp.SrcPort = uint16(or.SrcPort)
	}

	if or.DstPort > -1 {
		changed = true

		seq.Tcp.DstPort = uint16(or.DstPort)
		seq.Udp.DstPort = uint16(or.DstPort)
	}

	if or.TcpCooked > -1 {
		changed = true

		seq.Tcp.UseCookedSocket = utils.IntToBool(or.TcpCooked)
	}

	if or.TcpOneConn > -1 {
		changed = true

		seq.Tcp.UseOneConnection = utils.IntToBool(or.TcpOneConn)
	}

	if or.TcpSyn > -1 {
		changed = true

		seq.Tcp.Flags.Syn = utils.IntToBool(or.TcpSyn)
	}

	if or.TcpAck > -1 {
		changed = true

		seq.Tcp.Flags.Ack = utils.IntToBool(or.TcpAck)
	}

	if or.TcpPsh > -1 {
		changed = true

		seq.Tcp.Flags.Psh = utils.IntToBool(or.TcpPsh)
	}

	if or.TcpFin > -1 {
		changed = true

		seq.Tcp.Flags.Fin = utils.IntToBool(or.TcpFin)
	}

	if or.TcpRst > -1 {
		changed = true

		seq.Tcp.Flags.Rst = utils.IntToBool(or.TcpRst)
	}

	if or.TcpUrg > -1 {
		changed = true

		seq.Tcp.Flags.Urg = utils.IntToBool(or.TcpUrg)
	}

	if or.TcpEce > -1 {
		changed = true

		seq.Tcp.Flags.Ece = utils.IntToBool(or.TcpEce)
	}

	if or.TcpCwr > -1 {
		changed = true

		seq.Tcp.Flags.Cwr = utils.IntToBool(or.TcpCwr)
	}

	if or.IcmpCode > -1 {
		changed = true

		seq.Icmp.Code = uint8(or.IcmpCode)
	}

	if or.IcmpType > -1 {
		changed = true

		seq.Icmp.Type = uint8(or.IcmpType)
	}

	if or.PlMinLen > -1 || or.PlMaxLen > -1 || or.PlStatic > -1 || or.PlFile > -1 || or.PlString > -1 || len(or.PlExact) > 0 {
		changed = true

		var pl *config.Payload

		if len(seq.Payloads) > 0 {
			pl = &seq.Payloads[0]
		} else {
			pl = &config.Payload{}
		}

		if or.PlMinLen > -1 {
			pl.MinLen = uint16(or.PlMinLen)
		}

		if or.PlMaxLen > -1 {
			pl.MaxLen = uint16(or.PlMaxLen)
		}

		if or.PlStatic > -1 {
			pl.IsStatic = utils.IntToBool(or.PlStatic)
		}

		if or.PlFile > -1 {
			pl.IsFile = utils.IntToBool(or.PlFile)
		}

		if or.PlString > -1 {
			pl.IsString = utils.IntToBool(or.PlString)
		}

		if len(or.PlExact) > 0 {
			pl.Exact = &or.PlExact
		}

		if len(seq.Payloads) < 1 {
			seq.Payloads = append(seq.Payloads, *pl)
		}
	}

	if len(cfg.Sequences) < 1 && changed {
		cfg.Sequences = append(cfg.Sequences, *seq)
	}

	return len(cfg.Sequences) > 0
}
