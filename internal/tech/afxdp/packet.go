package afxdp

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

extern void* get_function(void* handle, const char* funcname);

int callSendPacket(void* f, void* xsk, void* pkt, int length, int batchSize) {
	int (*fn)(void*, void*, int, int) = (int (*)(void*, void*, int, int))f;

	return fn(xsk, pkt, length, batchSize);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func (c *Context) GetSendPacketFunc() error {
	cFunc := C.CString("SendPacket")
	defer C.free(unsafe.Pointer(cFunc))

	c.SendPacketFunc = C.get_function(c.Handle, cFunc)

	if c.SendPacketFunc == nil {
		return fmt.Errorf("failed to retrieve AF_XDP 'SendPacket' function: ret is nil")
	}

	return nil
}

func (c *Context) SendPacket(xsk unsafe.Pointer, data []byte, length int, batchSize int) error {
	ret := C.callSendPacket(c.SendPacketFunc, xsk, unsafe.Pointer(&data[0]), C.int(length), C.int(batchSize))

	if ret != 0 {
		return fmt.Errorf("failed to send packet: invalid return code (%d)", ret)
	}

	return nil
}
