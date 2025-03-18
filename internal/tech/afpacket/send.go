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

type sendPacketFuncType func(unsafe.Pointer, C.int, C.int) C.int

func (c *Context) GetSendPacketFunc() error {
	cFunc := C.CString("SendPacket")
	defer C.free(unsafe.Pointer(cFunc))

	c.SendPacketFunc = C.get_function(c.Handle, cFunc)

	if c.SendPacketFunc == nil {
		return fmt.Errorf("failed to retrieve AF_PACKET 'SendPacket' function: ret is nil")
	}

	return nil
}

func (c *Context) SendPacket(data []byte, length int, threadIdx int) error {
	fn := *(*sendPacketFuncType)(unsafe.Pointer(c.SendPacketFunc))

	ret := fn(unsafe.Pointer(&data[0]), C.int(length), C.int(threadIdx))

	if ret != 0 {
		return fmt.Errorf("failed to send packet: invalid return code (%d)", ret)
	}

	return nil
}
