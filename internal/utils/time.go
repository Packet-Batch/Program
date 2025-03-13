package utils

import "golang.org/x/sys/unix"

func GetBootTimeNS() (int64, error) {
	var ts unix.Timespec
	err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &ts)

	if err != nil {
		return 0, err
	}

	return ts.Nano(), nil
}
