package afxdp

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>


extern void* get_function(void* handle, const char* funcname);

void* callSetup(void* f, const char* dev, int queueId, int needWakeup, int sharedUmem, int forceSkb, int zeroCopy) {
	void* (*fn)(const char*, int, int, int, int, int) = (void* (*)(const char*, int, int, int, int, int))f;

	return (void*) fn(dev, queueId, needWakeup, sharedUmem, forceSkb, zeroCopy);
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

func (c *Context) Setup(dev string, queueId int, needWakeup bool, sharedUmem bool, forceSkb bool, zeroCopy bool) (unsafe.Pointer, error) {
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

	ret := C.callSetup(c.SetupFunc, cDev, C.int(queueId), dNeedWakeup, dSharedUmem, dForceSkb, dZeroCopy)

	if ret == nil {
		return nil, fmt.Errorf("failed to setup AF_XDP: Socket is null")
	}

	return unsafe.Pointer(ret), nil
}
