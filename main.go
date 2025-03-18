package main

import (
	"os"

	"github.com/Packet-Batch/Program/internal/cli"
	"github.com/Packet-Batch/Program/internal/config"
	"github.com/Packet-Batch/Program/internal/sequence"

	_ "github.com/Packet-Batch/Program/internal/tech"
)

func main() {
	// Load CLI.
	cli := &cli.Cli{}

	cli.Parse()

	// Load config and defaults.
	cfg := &config.Config{}

	cfg.LoadDefaults()

	err := cfg.LoadFromFs(cli.Cfg)

	if err != nil {
		cfg.DebugMsg(0, "[CFG] Failed to load config file (%s): %v", cli.Cfg, err)

		os.Exit(1)
	}

	if cfg.SaveCfg {
		err := cfg.SaveToFs(cli.Cfg, nil)

		if err != nil {
			cfg.DebugMsg(0, "[CFG] Failed to save config to file system (%s): %v", cli.Cfg, err)
		} else {
			cfg.DebugMsg(2, "[CFG] Saved config to file system (%s)...", cli.Cfg)
		}
	}

	// Override first sequence if needed.
	seqFound := sequence.CheckFirstSeqOverride(cfg, cli)

	if cli.List {
		cfg.List()

		os.Exit(0)
	}

	if !seqFound {
		cfg.DebugMsg(0, "[CFG] No sequences found after checking for CLI override! Exiting...")

		os.Exit(1)
	}

	for k, v := range cfg.Sequences {
		err := sequence.ProcessSeq(cfg, cli, &v, k+1)

		if err != nil {
			cfg.DebugMsg(1, "[SEQ] Failed to process sequence #%d due to error: %v", k+1, err)
		}
	}

	cfg.DebugMsg(1, "Sequences executed. Exiting...")

	os.Exit(0)
}
