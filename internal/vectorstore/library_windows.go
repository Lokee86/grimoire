//go:build windows

package vectorstore

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	statusOK             = 0
	statusBufferTooSmall = 2
)

type Library struct {
	path string
	dll  *windows.DLL

	lastError    *windows.Proc
	objectExists *windows.Proc
	ingest       *windows.Proc
	materialize  *windows.Proc
	open         *windows.Proc
	closeHandle  *windows.Proc
	info         *windows.Proc
	search       *windows.Proc

	mu     sync.Mutex
	closed bool
}

func Load(explicit string) (*Library, error) {
	path, err := FindLibrary(explicit)
	if err != nil {
		return nil, err
	}
	dll, err := windows.LoadDLL(path)
	if err != nil {
		return nil, fmt.Errorf("load vector engine %s: %w", path, err)
	}
	library := &Library{path: path, dll: dll}
	procedures := []struct {
		name   string
		target **windows.Proc
	}{
		{"gv_last_error_message", &library.lastError},
		{"gv_object_exists", &library.objectExists},
		{"gv_ingest_jsonl", &library.ingest},
		{"gv_materialize_jsonl", &library.materialize},
		{"gv_open_snapshot", &library.open},
		{"gv_close_snapshot", &library.closeHandle},
		{"gv_snapshot_info", &library.info},
		{"gv_search", &library.search},
	}
	for _, procedure := range procedures {
		found, findErr := dll.FindProc(procedure.name)
		if findErr != nil {
			_ = dll.Release()
			return nil, fmt.Errorf("load vector ABI procedure %s: %w", procedure.name, findErr)
		}
		*procedure.target = found
	}
	abi, findErr := dll.FindProc("gv_abi_version")
	if findErr != nil {
		_ = dll.Release()
		return nil, fmt.Errorf("load vector ABI version: %w", findErr)
	}
	version, _, _ := abi.Call()
	if version != 1 {
		_ = dll.Release()
		return nil, fmt.Errorf("unsupported vector ABI version %d", version)
	}
	return library, nil
}

func (library *Library) Path() string { return library.path }

func (library *Library) Close() error {
	library.mu.Lock()
	defer library.mu.Unlock()
	if library.closed {
		return nil
	}
	library.closed = true
	return library.dll.Release()
}

func (library *Library) ObjectExists(store, model, source string) (bool, error) {
	storeBytes, modelBytes, sourceBytes := []byte(store), []byte(model), []byte(source)
	var exists byte
	status, _, _ := library.objectExists.Call(
		bytePointer(storeBytes), uintptr(len(storeBytes)),
		bytePointer(modelBytes), uintptr(len(modelBytes)),
		bytePointer(sourceBytes), uintptr(len(sourceBytes)),
		uintptr(unsafe.Pointer(&exists)),
	)
	runtime.KeepAlive(storeBytes)
	runtime.KeepAlive(modelBytes)
	runtime.KeepAlive(sourceBytes)
	if err := library.status(status); err != nil {
		return false, err
	}
	return exists != 0, nil
}

func (library *Library) IngestJSONL(store, model, input string) (uint64, error) {
	storeBytes, modelBytes, inputBytes := []byte(store), []byte(model), []byte(input)
	var count uint64
	status, _, _ := library.ingest.Call(
		bytePointer(storeBytes), uintptr(len(storeBytes)),
		bytePointer(modelBytes), uintptr(len(modelBytes)),
		bytePointer(inputBytes), uintptr(len(inputBytes)),
		uintptr(unsafe.Pointer(&count)),
	)
	runtime.KeepAlive(storeBytes)
	runtime.KeepAlive(modelBytes)
	runtime.KeepAlive(inputBytes)
	return count, library.status(status)
}

func (library *Library) MaterializeJSONL(store, model, manifest, snapshot string) (string, error) {
	storeBytes, modelBytes := []byte(store), []byte(model)
	manifestBytes, snapshotBytes := []byte(manifest), []byte(snapshot)
	identity := make([]byte, 128)
	var identityLength uintptr
	status, _, _ := library.materialize.Call(
		bytePointer(storeBytes), uintptr(len(storeBytes)),
		bytePointer(modelBytes), uintptr(len(modelBytes)),
		bytePointer(manifestBytes), uintptr(len(manifestBytes)),
		bytePointer(snapshotBytes), uintptr(len(snapshotBytes)),
		bytePointer(identity), uintptr(len(identity)),
		uintptr(unsafe.Pointer(&identityLength)),
	)
	runtime.KeepAlive(storeBytes)
	runtime.KeepAlive(modelBytes)
	runtime.KeepAlive(manifestBytes)
	runtime.KeepAlive(snapshotBytes)
	if err := library.status(status); err != nil {
		return "", err
	}
	if identityLength > uintptr(len(identity)) {
		return "", ErrBufferTooSmall
	}
	return string(identity[:identityLength]), nil
}

func (library *Library) OpenSnapshot(path string) (*Engine, error) {
	pathBytes := []byte(path)
	var handle uint64
	status, _, _ := library.open.Call(
		bytePointer(pathBytes), uintptr(len(pathBytes)), uintptr(unsafe.Pointer(&handle)),
	)
	runtime.KeepAlive(pathBytes)
	if err := library.status(status); err != nil {
		return nil, err
	}
	return &Engine{library: library, handle: handle}, nil
}

func (library *Library) status(raw uintptr) error {
	status := int32(raw)
	if status == statusOK {
		return nil
	}
	if status == statusBufferTooSmall {
		return ErrBufferTooSmall
	}
	message := library.errorMessage()
	if message == "" {
		message = fmt.Sprintf("vector engine status %d", status)
	}
	return fmt.Errorf("%s", message)
}

func (library *Library) errorMessage() string {
	length, _, _ := library.lastError.Call(0, 0)
	if length == 0 {
		return ""
	}
	buffer := make([]byte, length)
	library.lastError.Call(bytePointer(buffer), uintptr(len(buffer)))
	return strings.TrimSpace(string(buffer))
}
