package tech

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

void* load_library(const char* libname) {
    return dlopen(libname, RTLD_LAZY);
}

void* get_function(void* handle, const char* funcname) {
    return dlsym(handle, funcname);
}
*/
import "C"

import (
	"github.com/Packet-Batch/Program/internal/tech/afpacket"
	"github.com/Packet-Batch/Program/internal/tech/afxdp"
	"github.com/Packet-Batch/Program/internal/tech/dpdk"
)

func Load(tech string) (interface{}, error) {
	switch tech {
	case "af_packet":
		return afpacket.New()

	case "dpdk":
		return dpdk.New()

	default:
		return afxdp.New()
	}
}
