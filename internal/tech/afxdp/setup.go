package afxdp

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

extern void* get_function(void* handle, const char* funcname);
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type setupFuncType func(*C.char, C.int, C.int, C.int, C.int, C.int, C.int) C.int

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

	fn := *(*setupFuncType)(unsafe.Pointer(c.SetupFunc))

	ret := fn(cDev, C.int(queueId), dNeedWakeup, dSharedUmem, dForceSkb, dZeroCopy, C.int(threads))

	if ret != 0 {
		return fmt.Errorf("failed to setup AF_XDP: invalid return code (%d)", ret)
	}

	return nil
}
