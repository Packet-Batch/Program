package dpdk

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

type setupFuncType func(*C.char, C.int) C.int

func (c *Context) GetSetupFunc() error {
	cFunc := C.CString("Setup")
	defer C.free(unsafe.Pointer(cFunc))

	c.SetupFunc = C.get_function(c.Handle, cFunc)

	if c.SetupFunc == nil {
		return fmt.Errorf("failed to retrieve DPDK 'Setup' function: ret is nil")
	}

	return nil
}

func (c *Context) Setup(dev string, threads int) error {
	cDev := C.CString(dev)
	defer C.free(unsafe.Pointer(cDev))

	fn := *(*setupFuncType)(unsafe.Pointer(c.SetupFunc))

	ret := fn(cDev, C.int(threads))

	if ret != 0 {
		return fmt.Errorf("failed to setup DPDK: invalid return code (%d)", ret)
	}

	return nil
}
