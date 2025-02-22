package afxdp

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

extern void* get_function(void* handle, const char* funcname);

int callSendPacket(void* f, void* pkt, int length, int threadIdx, int batchSize) {
	int (*fn)(void*, int, int, int) = (int (*)(void*, int, int, int))f;

	return fn(pkt, length, threadIdx, batchSize);
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

func (c *Context) SendPacket(data []byte, length int, threadIdx int, batchSize int) error {
	ret := C.callSendPacket(c.SendPacketFunc, unsafe.Pointer(&data[0]), C.int(length), C.int(threadIdx), C.int(batchSize))

	if ret != 0 {
		return fmt.Errorf("failed to send packet: invalid return code (%d)", ret)
	}

	return nil
}
