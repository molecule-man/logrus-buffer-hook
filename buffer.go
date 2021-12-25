package logrusbufferhook

import "io"

// Fixed-size ring buffer. If buffer becomes full the old data is overwritten.
type Buffer struct {
	data   []byte
	size   int
	cursor int
	isFull bool
}

func NewBuffer(size int) *Buffer {
	return &Buffer{
		size: size,
		data: make([]byte, size),
	}
}

func (b *Buffer) Write(buf []byte) (int, error) {
	n := len(buf)

	// we discard bytes from the beginning that will not fit into our
	// fixed-sized buffer
	if n > b.size {
		buf = buf[n-b.size:]
	}

	written := copy(b.data[b.cursor:], buf)
	if len(buf) > written {
		b.cursor = copy(b.data, buf[written:])
		b.isFull = true
	} else {
		b.cursor += written
	}

	return n, nil
}

func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	n, err := b.writeTo(w)
	if err != nil {
		return int64(n), err
	}

	// reset buffer
	b.cursor = 0
	b.isFull = false

	return int64(n), nil

}

func (b *Buffer) writeTo(w io.Writer) (int, error) {
	if !b.isFull {
		return w.Write(b.data[:b.cursor])
	}

	for i := b.cursor + 1; i < b.size; i++ {
		if b.data[i-1] == '\n' && b.data[i] != '\n' {
			n1, err := w.Write(b.data[i:])
			if err != nil {
				return n1, err
			}

			n2, err := w.Write(b.data[:b.cursor])
			return n1 + n2, err
		}
	}

	for i := 1; i < b.cursor; i++ {
		if b.data[i-1] == '\n' && b.data[i] != '\n' {
			return w.Write(b.data[i:b.cursor])
		}
	}

	return 0, nil
}

// Available returns number of bytes left in buffer. When number of available
// bytes becomes 0 then all consecutive writes will cause the oldest data to be
// overwritten.
func (b Buffer) Available() int {
	if b.isFull {
		return 0
	}

	return b.size - b.cursor
}
