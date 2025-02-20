package afpacket

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

type setupFuncType func(*C.char, C.int, C.int) C.int

func (c *Context) GetSetupFunc() error {
	cFunc := C.CString("Setup")
	defer C.free(unsafe.Pointer(cFunc))

	c.SetupFunc = C.get_function(c.Handle, cFunc)

	if c.SetupFunc == nil {
		return fmt.Errorf("failed to retrieve AF_PACKET 'Setup' function: ret is nil")
	}

	return nil
}

func (c *Context) Setup(dev string, cooked bool, threads int) error {
	cDev := C.CString(dev)
	defer C.free(unsafe.Pointer(cDev))

	dCooked := C.int(0)

	if cooked {
		dCooked = C.int(1)
	}

	fn := *(*setupFuncType)(unsafe.Pointer(c.SetupFunc))

	ret := fn(cDev, dCooked, C.int(threads))

	if ret != 0 {
		return fmt.Errorf("failed to setup AF_PACKET: invalid return code (%d)", ret)
	}

	return nil
}
