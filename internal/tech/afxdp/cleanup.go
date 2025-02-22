package afxdp

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

extern void* get_function(void* handle, const char* funcname);

int callCleanup(void* f, int threads) {
	int (*fn)(int) = (int (*)(int))f;

	return fn(threads);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

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
	ret := C.callCleanup(c.CleanupFunc, C.int(threads))

	if ret != 0 {
		return fmt.Errorf("failed to cleanup AF_XDP tech: invalid return code (%d)", ret)
	}

	return nil
}
