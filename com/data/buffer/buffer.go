package buffer

import (
	"bytes"
)

type Buffer struct {
	c chan *bytes.Buffer
}

func New(size int) (buff *Buffer) {
	return &Buffer{c: make(chan *bytes.Buffer, size)}
}

func (buff *Buffer) Get() (b *bytes.Buffer) {
	select {
	case b = <-buff.c:
	default:
		b = bytes.NewBuffer([]byte{})
	}
	return
}

func (buff *Buffer) Put(b *bytes.Buffer) {
	b.Reset()
	buff.c <- b
}
