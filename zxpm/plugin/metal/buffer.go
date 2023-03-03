package metal

import (
	"io"
	"sync"
)

type SyncBuffer struct {
	sync.Mutex
	w io.Writer
}

func NewSyncBuffer(w io.Writer) *SyncBuffer {
	return &SyncBuffer{w: w}
}

func (b *SyncBuffer) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()
	return b.w.Write(p)
}
