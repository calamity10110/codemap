//go:build windows

package scanner

import (
	"fmt"
	"syscall"
	"unsafe"
)

// loadLibrary loads a DLL on Windows
func loadLibrary(path string) (uintptr, error) {
	handle, err := syscall.LoadLibrary(path)
	if err != nil {
		return 0, err
	}
	return uintptr(handle), nil
}

// getLanguageFunc gets the tree_sitter_<lang> function from the DLL
func getLanguageFunc(lib uintptr, lang string) (func() unsafe.Pointer, error) {
	procName := fmt.Sprintf("tree_sitter_%s", lang)
	proc, err := syscall.GetProcAddress(syscall.Handle(lib), procName)
	if err != nil {
		return nil, fmt.Errorf("GetProcAddress %s: %w", procName, err)
	}

	// Create a wrapper function that calls the proc
	fn := func() unsafe.Pointer {
		ret, _, _ := syscall.SyscallN(proc)
		return unsafe.Pointer(ret)
	}
	return fn, nil
}
