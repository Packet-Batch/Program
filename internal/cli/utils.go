package cli

import "flag"

func StringOpt(opt *string, short string, long string, def string, usage string) {
	if len(short) > 0 {
		flag.StringVar(opt, short, def, usage)
	}

	if len(long) > 0 {
		flag.StringVar(opt, long, def, usage)
	}
}

func IntOpt(opt *int, short string, long string, def int, usage string) {
	if len(short) > 0 {
		flag.IntVar(opt, short, def, usage)
	}

	if len(long) > 0 {
		flag.IntVar(opt, long, def, usage)
	}
}

func BoolOpt(opt *bool, short string, long string, def bool, usage string) {
	if len(short) > 0 {
		flag.BoolVar(opt, short, def, usage)
	}

	if len(long) > 0 {
		flag.BoolVar(opt, long, def, usage)
	}
}

func Int64Opt(opt *int64, short string, long string, def int64, usage string) {
	if len(short) > 0 {
		flag.Int64Var(opt, short, def, usage)
	}

	if len(long) > 0 {
		flag.Int64Var(opt, long, def, usage)
	}
}
