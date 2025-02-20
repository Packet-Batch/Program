package dpdk

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

extern void* load_library(const char* libname);
extern void* get_function(void* handle, const char* funcname);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type Context struct {
	Handle unsafe.Pointer

	SetupFunc      unsafe.Pointer
	CleanupFunc    unsafe.Pointer
	SendPacketFunc unsafe.Pointer
}

func New() (*Context, error) {
	ctx := &Context{}

	cLib := C.CString("libtechdpdk.so")

	defer C.free(unsafe.Pointer(cLib))

	ctx.Handle = C.load_library(cLib)

	if ctx.Handle == nil {
		return nil, fmt.Errorf("failed to load DPDK library: handle is nil")
	}

	err := ctx.GetSetupFunc()

	if err != nil {
		return nil, err
	}

	err = ctx.GetCleanupFunc()

	if err != nil {
		return nil, err
	}

	err = ctx.GetSendPacketFunc()

	if err != nil {
		return nil, err
	}

	return ctx, nil
}
