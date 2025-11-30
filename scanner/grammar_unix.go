//go:build !windows

package scanner

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
)

// loadLibrary loads a shared library on Unix systems
func loadLibrary(path string) (uintptr, error) {
	lib, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return 0, err
	}
	return lib, nil
}

// getLanguageFunc gets the tree_sitter_<lang> function from the library
func getLanguageFunc(lib uintptr, lang string) (func() unsafe.Pointer, error) {
	var langFunc func() unsafe.Pointer
	purego.RegisterLibFunc(&langFunc, lib, fmt.Sprintf("tree_sitter_%s", lang))
	return langFunc, nil
}
