package config

import (
	"fmt"
	"strconv"

	"github.com/Packet-Batch/Program/internal/utils"
)

func (cfg *Config) List() {
	fmt.Printf("Listing settings...\n\n")

	fmt.Printf("Debug Settings\n")
	fmt.Printf("\tVerbose => %s\n", strconv.Itoa(cfg.Debug.Verbose))

	logDir := "N/A"

	if cfg.Debug.LogDir != nil {
		logDir = *cfg.Debug.LogDir
	}

	fmt.Printf("\tLog Directory => %s\n", logDir)

	fmt.Printf("\nGeneral Settings\n")
	dev := "N/A"

	if len(cfg.Interface) > 0 {
		dev = cfg.Interface
	}

	fmt.Printf("\tInterface => %s\n", dev)
	fmt.Printf("\tSave Config => %s\n", strconv.FormatBool(cfg.SaveCfg))

	fmt.Printf("\nSequences\n")
	if len(cfg.Sequences) > 0 {
		for k, v := range cfg.Sequences {
			fmt.Printf("\t#%s\n", strconv.Itoa(k+1))

			dev := "N/A"

			if len(v.Interface) > 0 {
				dev = v.Interface
			}

			fmt.Printf("\t\tTech => %s\n", v.Tech)
			fmt.Printf("\t\tInterface => %s\n", dev)
			fmt.Printf("\t\tBlock => %s\n", strconv.FormatBool(v.Block))
			fmt.Printf("\t\tTrack => %s\n", strconv.FormatBool(v.Track))
			fmt.Printf("\t\tPPS => %s\n", strconv.FormatUint(v.Pps, 10))
			fmt.Printf("\t\tBPS => %s\n", strconv.FormatUint(v.Bps, 10))
			fmt.Printf("\t\tTime => %s\n", strconv.Itoa(v.Time))
			fmt.Printf("\t\tDelay => %s\n", strconv.FormatUint(v.Delay, 10))
			fmt.Printf("\t\tThreads => %s\n", strconv.Itoa(int(v.Threads)))

			fmt.Printf("\n\t\tEthernet\n")

			ethSrcMac := "AUTO"

			if len(v.Eth.SrcMac) > 0 {
				ethSrcMac = v.Eth.SrcMac
			}

			fmt.Printf("\t\t\tSrc MAC => %s\n", ethSrcMac)

			ethDstMac := "AUTO"

			if len(v.Eth.DstMac) > 0 {
				ethDstMac = v.Eth.DstMac
			}

			fmt.Printf("\t\t\tDst MAC => %s\n", ethDstMac)

			fmt.Printf("\n\t\tIPv4\n")

			fmt.Printf("\t\t\tProtocol => %s\n", v.Ip4.Protocol)

			srcIp := "AUTO"

			if len(v.Ip4.SrcIp) > 0 {
				srcIp = v.Ip4.SrcIp
			}

			fmt.Printf("\t\t\tSrc IP => %s\n", srcIp)

			fmt.Printf("\t\t\tSrc Ranges\n")
			if len(v.Ip4.SrcIpRanges) > 0 {
				for _, v2 := range v.Ip4.SrcIpRanges {
					fmt.Printf("\t\t\t\t- %s\n", v2)
				}
			} else {
				fmt.Printf("\t\t\t\t- None\n")
			}

			fmt.Printf("\t\t\tDst IP => %s\n", v.Ip4.DstIp)
			fmt.Printf("\t\t\tMin TTL => %s\n", strconv.Itoa(int(v.Ip4.MinTtl)))
			fmt.Printf("\t\t\tMax TTL => %s\n", strconv.Itoa(int(v.Ip4.MaxTtl)))
			fmt.Printf("\t\t\tMin ID => %s\n", strconv.Itoa(int(v.Ip4.MinId)))
			fmt.Printf("\t\t\tMax ID => %s\n", strconv.Itoa(int(v.Ip4.MaxId)))
			fmt.Printf("\t\t\tChecksum => %s\n", strconv.FormatBool(v.Ip4.Csum))

			fmt.Printf("\n\t\tTCP\n")

			fmt.Printf("\t\t\tSrc Port => %s\n", strconv.Itoa(int(v.Tcp.SrcPort)))
			fmt.Printf("\t\t\tDst Port => %s\n", strconv.Itoa(int(v.Tcp.DstPort)))
			fmt.Printf("\t\t\tUse Cooked Socket => %s\n", strconv.FormatBool(v.Tcp.UseCookedSocket))
			fmt.Printf("\t\t\tUse One Connection => %s\n", strconv.FormatBool(v.Tcp.UseOneConnection))
			fmt.Printf("\t\t\tChecksum => %s\n", strconv.FormatBool(v.Tcp.Csum))

			fmt.Printf("\n\t\t\tFlags\n")

			fmt.Printf("\t\t\t\tSYN => %s\n", strconv.FormatBool(v.Tcp.Flags.Syn))
			fmt.Printf("\t\t\t\tACK => %s\n", strconv.FormatBool(v.Tcp.Flags.Ack))
			fmt.Printf("\t\t\t\tPSH => %s\n", strconv.FormatBool(v.Tcp.Flags.Psh))
			fmt.Printf("\t\t\t\tFIN => %s\n", strconv.FormatBool(v.Tcp.Flags.Fin))
			fmt.Printf("\t\t\t\tRST => %s\n", strconv.FormatBool(v.Tcp.Flags.Rst))
			fmt.Printf("\t\t\t\tURG => %s\n", strconv.FormatBool(v.Tcp.Flags.Urg))
			fmt.Printf("\t\t\t\tECE => %s\n", strconv.FormatBool(v.Tcp.Flags.Ece))
			fmt.Printf("\t\t\t\tCWR => %s\n", strconv.FormatBool(v.Tcp.Flags.Cwr))

			fmt.Printf("\n\t\tUDP\n")

			fmt.Printf("\t\t\tSrc Port => %s\n", strconv.Itoa(int(v.Udp.SrcPort)))
			fmt.Printf("\t\t\tDst Port => %s\n", strconv.Itoa(int(v.Udp.DstPort)))
			fmt.Printf("\t\t\tChecksum => %s\n", strconv.FormatBool(v.Udp.Csum))

			fmt.Printf("\n\t\tICMP\n")

			fmt.Printf("\t\t\tCode => %s\n", strconv.Itoa(int(v.Icmp.Code)))
			fmt.Printf("\t\t\tType => %s\n", strconv.Itoa(int(v.Icmp.Type)))
			fmt.Printf("\t\t\tChecksum => %s\n", strconv.FormatBool(v.Icmp.Csum))

			fmt.Printf("\n\t\tPayloads\n")
			if len(v.Payloads) > 0 {
				for k2, v2 := range v.Payloads {
					fmt.Printf("\t\t\t#%s\n", strconv.Itoa(k2+1))

					fmt.Printf("\t\t\t\tMin Length => %s\n", strconv.Itoa(int(v2.MinLen)))
					fmt.Printf("\t\t\t\tMax Length => %s\n", strconv.Itoa(int(v2.MaxLen)))
					fmt.Printf("\t\t\t\tIs Static => %s\n", strconv.FormatBool(v2.IsStatic))
					fmt.Printf("\t\t\t\tIs File => %s\n", strconv.FormatBool(v2.IsFile))
					fmt.Printf("\t\t\t\tIs String => %s\n", strconv.FormatBool(v2.IsString))

					exact := "N/A"

					if len(v2.Exact) > 0 {
						exact = v2.Exact
					}

					fmt.Printf("\t\t\t\tExact => %s\n", exact)
				}
			} else {
				fmt.Printf("\t\t\t- None\n")
			}

		}
	} else {
		fmt.Printf("\t- None\n")
	}
}

func (cfg *Config) DebugMsg(req_lvl int, msg string, args ...interface{}) {
	f_msg := fmt.Sprintf(msg, args...)

	utils.DebugMsg(req_lvl, cfg.Debug.Verbose, cfg.Debug.LogDir, f_msg)
}
