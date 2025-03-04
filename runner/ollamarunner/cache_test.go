package ollamarunner

import (
	"errors"
	"fmt"
	"image"
	"testing"
	"time"

	"github.com/ollama/ollama/ml"
)

func TestCountCommon(t *testing.T) {
	imgA := image.NewRGBA(image.Rect(0, 0, 100, 100))
	imgB := image.NewRGBA(image.Rect(0, 0, 50, 50))
	imgC := image.NewRGBA(image.Rect(50, 50, 100, 100))

	tests := []struct {
		name     string
		t1       []input
		t2       []input
		expected int32
	}{
		{
			name:     "Equal",
			t1:       []input{{token: 1}, {token: 2}, {token: 3}},
			t2:       []input{{token: 1}, {token: 2}, {token: 3}},
			expected: 3,
		},
		{
			name:     "Prefix",
			t1:       []input{{token: 1}},
			t2:       []input{{token: 1}, {token: 2}, {token: 3}},
			expected: 1,
		},
		{
			name:     "Image Prefix",
			t1:       []input{{image: imgA}},
			t2:       []input{{image: imgA}, {image: imgB}, {image: imgC}},
			expected: 1,
		},
		{
			name:     "Mixed",
			t1:       []input{{token: 1}, {image: imgA}},
			t2:       []input{{token: 1}, {image: imgA}, {token: 5}},
			expected: 2,
		},
		{
			name:     "Empty",
			t1:       []input{},
			t2:       []input{{token: 1}, {token: 2}, {token: 3}},
			expected: 0,
		},
		{
			name:     "Both Empty",
			t1:       []input{},
			t2:       []input{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countCommonPrefix(tt.t1, tt.t2)
			if result != tt.expected {
				t.Errorf("countCommonPrefix(%v, %v): have %v; want %v", tt.t1, tt.t2, result, tt.expected)
			}
		})
	}
}

func TestFindCacheSlot(t *testing.T) {
	type expected struct {
		result int
		len    int32
	}

	tests := []struct {
		name    string
		cache   InputCache
		prompt  []input
		longest expected
		best    expected
	}{
		{
			name: "Empty",
			cache: InputCache{slots: []InputCacheSlot{
				{
					Id:       0,
					Inputs:   []input{},
					InUse:    false,
					lastUsed: time.Time{},
				},
				{
					Id:       1,
					Inputs:   []input{},
					InUse:    false,
					lastUsed: time.Time{},
				},
			}},
			prompt:  []input{{token: 1}},
			longest: expected{result: 0, len: 0},
			best:    expected{result: 0, len: 0},
		},
		{
			name: "Extend",
			cache: InputCache{slots: []InputCacheSlot{
				{
					Id:       0,
					Inputs:   []input{{token: 1}},
					InUse:    false,
					lastUsed: time.Now().Add(-time.Second),
				},
				{
					Id:       1,
					Inputs:   []input{{token: 1}, {token: 2}},
					InUse:    false,
					lastUsed: time.Now().Add(-2 * time.Second),
				},
			}},
			prompt:  []input{{token: 1}, {token: 2}},
			longest: expected{result: 1, len: 2},
			best:    expected{result: 1, len: 2},
		},
		{
			name: "New",
			cache: InputCache{slots: []InputCacheSlot{
				{
					Id:       0,
					Inputs:   []input{{token: 1}, {token: 2}},
					InUse:    false,
					lastUsed: time.Now().Add(-time.Second),
				},
				{
					Id:       1,
					Inputs:   []input{},
					InUse:    false,
					lastUsed: time.Time{},
				},
			}},
			prompt:  []input{{token: 2}},
			longest: expected{result: 0, len: 0},
			best:    expected{result: 1, len: 0},
		},
		{
			name: "Fork",
			cache: InputCache{
				slots: []InputCacheSlot{
					{
						Id:       0,
						Inputs:   []input{{token: 1}, {token: 2}},
						InUse:    false,
						lastUsed: time.Now().Add(-time.Second),
					},
					{
						Id:       1,
						Inputs:   []input{},
						InUse:    false,
						lastUsed: time.Time{},
					},
				},
			},
			prompt:  []input{{token: 1}},
			longest: expected{result: 0, len: 1},
			best:    expected{result: 1, len: 1},
		},
		{
			name: "Evict",
			cache: InputCache{slots: []InputCacheSlot{
				{
					Id:       0,
					Inputs:   []input{{token: 1}},
					InUse:    false,
					lastUsed: time.Now().Add(-time.Second),
				},
				{
					Id:       1,
					Inputs:   []input{{token: 1}, {token: 2}},
					InUse:    false,
					lastUsed: time.Now().Add(-2 * time.Second),
				},
			}},
			prompt:  []input{{token: 2}, {token: 3}},
			longest: expected{result: 0, len: 0},
			best:    expected{result: 1, len: 0},
		},
		{
			name: "In use",
			cache: InputCache{slots: []InputCacheSlot{
				{
					Id:       0,
					Inputs:   []input{{token: 1}, {token: 2}},
					InUse:    true,
					lastUsed: time.Now().Add(-time.Second),
				},
				{
					Id:       1,
					Inputs:   []input{{token: 1}},
					InUse:    false,
					lastUsed: time.Now().Add(-2 * time.Second),
				},
			}},
			prompt:  []input{{token: 1}, {token: 2}},
			longest: expected{result: 1, len: 1},
			best:    expected{result: 1, len: 2},
		},
	}

	for _, tt := range tests {
		t.Run("Longest-"+tt.name, func(t *testing.T) {
			result, resultLen, err := tt.cache.findLongestCacheSlot(tt.prompt)
			if err != nil {
				t.Errorf("findLongestCacheSlot: err %v", err)
			} else if result.Id != tt.longest.result || resultLen != tt.longest.len {
				t.Errorf("findLongestCacheSlot: slot have %v, want %v len have %v, want %v",
					result.Id, tt.longest.result, resultLen, tt.longest.len)
			}
		})
	}

	for _, tt := range tests {
		t.Run("Best-"+tt.name, func(t *testing.T) {
			result, resultLen, err := tt.cache.findBestCacheSlot(tt.prompt)
			if err != nil {
				t.Errorf("findBestCacheSlot: err %v", err)
			} else if result.Id != tt.best.result || resultLen != tt.best.len {
				t.Errorf("findBestCacheSlot: slot have %v, want %v len have %v, want %v",
					result.Id, tt.best.result, resultLen, tt.best.len)
			}
		})
	}
}

func TestShiftDiscard(t *testing.T) {
	tests := []struct {
		name     string
		numCtx   int32
		numKeep  int32
		inputLen int32
		expected int32
	}{
		{
			name:     "Shift",
			numCtx:   2048,
			numKeep:  5,
			inputLen: 2048,
			expected: 1021,
		},
		{
			name:     "Max Keep",
			numCtx:   2048,
			numKeep:  2047,
			inputLen: 2048,
			expected: 1,
		},
		{
			name:     "No Keep",
			numCtx:   2048,
			numKeep:  0,
			inputLen: 2048,
			expected: 1024,
		},
		{
			name:     "Truncate",
			numCtx:   2048,
			numKeep:  5,
			inputLen: 5000,
			expected: 3973,
		},
		{
			name:     "Truncate Keep",
			numCtx:   2048,
			numKeep:  2047,
			inputLen: 5000,
			expected: 2953,
		},
		{
			name:     "No Op",
			numCtx:   2048,
			numKeep:  5,
			inputLen: 512,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := InputCache{numCtx: tt.numCtx}
			result := c.ShiftDiscard(tt.inputLen, tt.numKeep)
			if result != tt.expected {
				t.Errorf("shiftDiscard(ctx: %v, keep: %v input: %v): have %v; want %v", tt.numCtx, tt.numKeep, tt.inputLen, result, tt.expected)
			}
		})
	}
}

// Mock implementation of the Cache interface
type mockCache struct {
	shouldFail bool
}

// Implement only the methods needed for the test
func (m *mockCache) Remove(seq int, beginIndex, endIndex int32) error {
	if m.shouldFail {
		return fmt.Errorf("mock cache removal error")
	}
	return nil
}

// Stub implementations for other interface methods
func (m *mockCache) SetLayer(layer int)                                               {}
func (m *mockCache) Get(ctx ml.Context) (ml.Tensor, ml.Tensor, ml.Tensor)             { return nil, nil, nil }
func (m *mockCache) Put(ctx ml.Context, key, value ml.Tensor)                         {}
func (m *mockCache) Init(backend ml.Backend, dtype ml.DType, capacity int32)          {}
func (m *mockCache) Close()                                                           {}
func (m *mockCache) StartForward(ctx ml.Context, positions []int32, seqs []int) error { return nil }
func (m *mockCache) CopyPrefix(srcSeq, dstSeq int, len int32)                         {}

func TestShiftCacheSlot(t *testing.T) {
	tests := []struct {
		name          string
		numCtx        int32
		inputs        []input
		numKeep       int32
		cacheErr      bool
		wantErr       any
		wantInputsLen int
	}{
		{
			name:          "Normal shift",
			numCtx:        10,
			inputs:        []input{{token: 1}, {token: 2}, {token: 3}, {token: 4}, {token: 5}, {token: 6}, {token: 7}, {token: 8}, {token: 9}, {token: 10}},
			numKeep:       2,
			cacheErr:      false, // No error
			wantErr:       nil,
			wantInputsLen: 6, // After discarding 4 tokens
		},
		{
			name:          "Cache removal fails",
			numCtx:        10,
			inputs:        []input{{token: 1}, {token: 2}, {token: 3}, {token: 4}, {token: 5}, {token: 6}, {token: 7}, {token: 8}, {token: 9}, {token: 10}},
			numKeep:       2,
			cacheErr:      true,
			wantErr:       &ErrReprocessInputs{},
			wantInputsLen: 0, // Original inputs should be cleared
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCache{shouldFail: tt.cacheErr}
			c := InputCache{
				numCtx: tt.numCtx,
				cache:  mock,
			}
			slot := &InputCacheSlot{
				Id:     123,
				Inputs: make([]input, len(tt.inputs)),
			}
			copy(slot.Inputs, tt.inputs)

			err := c.ShiftCacheSlot(slot, tt.numKeep)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}

				if !errors.As(err, &tt.wantErr) {
					t.Errorf("Expected error of type %T but got %T: %v", tt.wantErr, err, err)
				}

				if errReproc, ok := err.(*ErrReprocessInputs); ok {
					if errReproc.SlotId != slot.Id {
						t.Errorf("ErrReprocessInputs has wrong SlotId: got %v, want %v", errReproc.SlotId, slot.Id)
					}
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(slot.Inputs) != tt.wantInputsLen {
				t.Errorf("Slot inputs length after operation: got %v, want %v", len(slot.Inputs), tt.wantInputsLen)
			}
		})
	}
}
