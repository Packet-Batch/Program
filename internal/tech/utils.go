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
