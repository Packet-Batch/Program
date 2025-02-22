package afxdp

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

extern void* get_function(void* handle, const char* funcname);

int callSetup(void* f, const char* dev, int queueId, int needWakeup, int sharedUmem, int forceSkb, int zeroCopy, int threads) {
	int (*fn)(const char*, int, int, int, int, int, int) = (int (*)(const char*, int, int, int, int, int, int))f;

	return fn(dev, queueId, needWakeup, sharedUmem, forceSkb, zeroCopy, threads);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func (c *Context) GetSetupFunc() error {
	cFunc := C.CString("Setup")
	defer C.free(unsafe.Pointer(cFunc))

	c.SetupFunc = C.get_function(c.Handle, cFunc)

	if c.SetupFunc == nil {
		return fmt.Errorf("failed to retrieve AF_XDP 'Setup' function: ret is nil")
	}

	return nil
}

func (c *Context) Setup(dev string, queueId int, needWakeup bool, sharedUmem bool, forceSkb bool, zeroCopy bool, threads int) error {
	cDev := C.CString(dev)
	defer C.free(unsafe.Pointer(cDev))

	dNeedWakeup := C.int(0)

	if needWakeup {
		dNeedWakeup = C.int(1)
	}

	dSharedUmem := C.int(0)

	if sharedUmem {
		dSharedUmem = C.int(1)
	}

	dForceSkb := C.int(0)

	if forceSkb {
		dForceSkb = C.int(1)
	}

	dZeroCopy := C.int(0)

	if zeroCopy {
		dZeroCopy = C.int(1)
	}

	ret := C.callSetup(c.SetupFunc, cDev, C.int(queueId), dNeedWakeup, dSharedUmem, dForceSkb, dZeroCopy, C.int(threads))

	if ret != 0 {
		return fmt.Errorf("failed to setup AF_XDP: invalid return code (%d)", ret)
	}

	return nil
}
