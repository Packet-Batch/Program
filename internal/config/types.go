package config

type Debug struct {
	Verbose int     `json:"Verbose"`
	LogDir  *string `json:"LogDir"`
}

type Eth struct {
	SrcMac string `json:"SrcMac"`
	DstMac string `json:"DstMac"`
}

type Ip4 struct {
	Protocol string `json:"Protocol"`

	SrcIp       string   `json:"SrcIp"`
	SrcIpRanges []string `json:"SrcIpRanges"`

	DstIp string `json:"DstIp"`

	Tos uint8 `json:"Tos"`

	MinTtl uint8 `json:"MinTtl"`
	MaxTtl uint8 `json:"MaxTtl"`

	MinId uint16 `json:"MinId"`
	MaxId uint16 `json:"MaxId"`

	Csum bool `json:"DoCsum"`
}

type TcpFlags struct {
	Syn bool `json:"Syn"`
	Ack bool `json:"Ack"`
	Psh bool `json:"Psh"`
	Fin bool `json:"Fin"`
	Rst bool `json:"Rst"`
	Urg bool `json:"Urg"`
	Ece bool `json:"Ece"`
	Cwr bool `json:"Cwr"`
}

type Tcp struct {
	SrcPort uint16 `json:"SrcPort"`
	DstPort uint16 `json:"DstPort"`

	UseCookedSocket  bool `json:"UseCookedSocket"`
	UseOneConnection bool `json:"UseOneConnection"`

	Flags TcpFlags `json:"Flags"`

	Csum bool `json:"Csum"`
}

type Udp struct {
	SrcPort uint16 `json:"SrcPort"`
	DstPort uint16 `json:"DstPort"`

	Csum bool `json:"Csum"`
}

type Icmp struct {
	Code uint8 `json:"Code"`
	Type uint8 `json:"Type"`

	Csum bool `json:"Csum"`
}

type Payload struct {
	MinLen uint16 `json:"MinLen"`
	MaxLen uint16 `json:"MaxLen"`

	IsStatic bool `json:"IsStatic"`

	IsFile   bool `json:"IsFile"`
	IsString bool `json:"IsString"`

	Exact string `json:"Exact"`
}

type Sequence struct {
	Tech string `json:"Tech"`

	Interface string `json:"Interface"`

	Block bool `json:"Block"`
	Track bool `json:"Track"`

	MaxPkts  uint64 `json:"MaxPkts"`
	MaxBytes uint64 `json:"MaxBytes"`

	Pps uint64 `json:"Pps"`
	Bps uint64 `json:"Bps"`

	Time    int    `json:"Time"`
	Delay   uint64 `json:"Delay"`
	Threads uint8  `json:"Threads"`

	Includes []string `json:"Includes"`

	Eth  Eth  `json:"Eth"`
	Ip4  Ip4  `json:"Ip4"`
	Tcp  Tcp  `json:"Tcp"`
	Udp  Udp  `json:"Udp"`
	Icmp Icmp `json:"Icmp"`

	Payloads []Payload `json:"Payloads"`
}

type Config struct {
	Debug     Debug  `json:"Debug"`
	Interface string `json:"Interface"`
	SaveCfg   bool   `json:"SaveCfg"`

	Sequences []Sequence `json:"Sequences"`
}

func (cfg *Config) LoadDefaults() {
	cfg.Debug.Verbose = 1
}
