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

type cleanupFuncType func(C.int) C.int

func (c *Context) GetCleanupFunc() error {
	cFunc := C.CString("Cleanup")
	defer C.free(unsafe.Pointer(cFunc))

	c.CleanupFunc = C.get_function(c.Handle, cFunc)

	if c.CleanupFunc == nil {
		return fmt.Errorf("failed to retrieve AF_XDP 'Cleanup' function: ret is nil")
	}

	return nil
}

func (c *Context) Cleanup(threads int) error {
	fn := *(*cleanupFuncType)(unsafe.Pointer(c.CleanupFunc))

	ret := fn(C.int(threads))

	if ret != 0 {
		return fmt.Errorf("failed to cleanup AF_XDP tech: invalid return code (%d)", ret)
	}

	return nil
}
