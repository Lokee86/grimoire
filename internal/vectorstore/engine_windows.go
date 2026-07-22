//go:build windows

package vectorstore

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

type Engine struct {
	library *Library
	handle  uint64

	mu     sync.RWMutex
	closed bool
}

type ffiResult struct {
	IDOffset uint64
	IDLength uint32
	Reserved uint32
	Score    float32
	Padding  uint32
	Index    uint64
}

func (engine *Engine) Info() (Info, error) {
	engine.mu.RLock()
	defer engine.mu.RUnlock()
	if engine.closed {
		return Info{}, errors.New("vector snapshot is closed")
	}
	model := make([]byte, 1024)
	var dimensions uint32
	var count uint64
	var modelLength uintptr
	status, _, _ := engine.library.info.Call(
		uintptr(engine.handle),
		uintptr(unsafe.Pointer(&dimensions)),
		uintptr(unsafe.Pointer(&count)),
		bytePointer(model), uintptr(len(model)),
		uintptr(unsafe.Pointer(&modelLength)),
	)
	runtime.KeepAlive(model)
	if err := engine.library.status(status); err != nil {
		return Info{}, err
	}
	if modelLength > uintptr(len(model)) {
		return Info{}, ErrBufferTooSmall
	}
	return Info{Model: string(model[:modelLength]), Dimensions: int(dimensions), Count: int(count)}, nil
}

func (engine *Engine) Search(query []float32, topK int) ([]Hit, error) {
	engine.mu.RLock()
	defer engine.mu.RUnlock()
	if engine.closed {
		return nil, errors.New("vector snapshot is closed")
	}
	if len(query) == 0 || topK <= 0 {
		return nil, errors.New("query and positive topK are required")
	}

	results := make([]ffiResult, topK)
	ids := make([]byte, max(1024, topK*128))
	for attempts := 0; attempts < 2; attempts++ {
		var count uintptr
		var idsLength uintptr
		status, _, _ := engine.library.search.Call(
			uintptr(engine.handle),
			uintptr(unsafe.Pointer(unsafe.SliceData(query))), uintptr(len(query)), uintptr(topK),
			uintptr(unsafe.Pointer(unsafe.SliceData(results))), uintptr(len(results)),
			bytePointer(ids), uintptr(len(ids)),
			uintptr(unsafe.Pointer(&count)), uintptr(unsafe.Pointer(&idsLength)),
		)
		runtime.KeepAlive(query)
		runtime.KeepAlive(results)
		runtime.KeepAlive(ids)
		if int32(status) == statusBufferTooSmall {
			if count > uintptr(len(results)) {
				results = make([]ffiResult, count)
			}
			if idsLength > uintptr(len(ids)) {
				ids = make([]byte, idsLength)
			}
			continue
		}
		if err := engine.library.status(status); err != nil {
			return nil, err
		}
		if count > uintptr(len(results)) || idsLength > uintptr(len(ids)) {
			return nil, ErrBufferTooSmall
		}
		hits := make([]Hit, int(count))
		for index, result := range results[:count] {
			start := int(result.IDOffset)
			end := start + int(result.IDLength)
			if start < 0 || end < start || end > int(idsLength) {
				return nil, fmt.Errorf("vector engine returned invalid id bounds %d:%d", start, end)
			}
			hits[index] = Hit{ID: string(ids[start:end]), Score: result.Score, Index: result.Index}
		}
		return hits, nil
	}
	return nil, ErrBufferTooSmall
}

func (engine *Engine) Close() error {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	if engine.closed {
		return nil
	}
	status, _, _ := engine.library.closeHandle.Call(uintptr(engine.handle))
	if err := engine.library.status(status); err != nil {
		return err
	}
	engine.closed = true
	return nil
}
