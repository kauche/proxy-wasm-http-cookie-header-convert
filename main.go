package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"

	"github.com/kauche/proxy-wasm-http-cookie-header-convert/internal"
)

func main() {
	proxywasm.SetVMContext(&internal.VmContext{})
}
