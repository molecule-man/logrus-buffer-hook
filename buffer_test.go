package logrusbufferhook

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestBuffer(t *testing.T) {
	out := &bytes.Buffer{}
	buf := NewBuffer(16)
	buf.Write([]byte("123456789\n"))
	buf.Write([]byte("12345"))
	buf.WriteTo(out)

	expected := "123456789\n12345"
	got := out.String()
	if got != expected {
		t.Errorf("expected: %q, got: %q", expected, got)
	}

	out.Reset()
	buf.Write([]byte("12345\n789\n"))
	buf.Write([]byte("overflown"))
	buf.WriteTo(out)

	expected = "789\noverflown"
	got = out.String()
	if got != expected {
		t.Errorf("expected: %q, got: %q", expected, got)
	}

	out.Reset()
	buf.Write([]byte("0123456789"))
	buf.Write([]byte("xxxxxxx\nabcd"))
	buf.WriteTo(out)

	expected = "abcd"
	got = out.String()
	if got != expected {
		t.Errorf("expected: %q, got: %q", expected, got)
	}
}

type bufInterface interface {
	io.Writer
	io.WriterTo
}

func benchmarkWrite(b *testing.B, buf bufInterface) {
	b.Helper()
	content := []byte(strings.Repeat("x", 100) + "\n")
	for i := 0; i < b.N; i++ {
		buf.Write(content)
	}
}

func benchmarkFlushIncompleteBuf(b *testing.B, buf bufInterface) {
	b.Helper()
	content := []byte(strings.Repeat("x", 200) + "\n")
	for i := 0; i < b.N; i++ {
		buf.Write(content)
		buf.WriteTo(io.Discard)
	}
}

func benchmarkFlushFullBuf(b *testing.B, buf bufInterface) {
	b.Helper()
	line := strings.Repeat("x", 101) + "\n"
	lineB := []byte(line)
	initialContent := []byte(strings.Repeat(line, 5))

	for linesNum := 1; linesNum < 5; linesNum++ {
		b.Run(fmt.Sprintf("input_size_%d", linesNum), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf.Write(initialContent)
				for l := 0; l < linesNum; l++ {
					buf.Write(lineB)
				}
				buf.WriteTo(io.Discard)
			}

		})
	}
}

func BenchmarkRingWrite(b *testing.B) {
	benchmarkWrite(b, NewBuffer(512))
}

func BenchmarkRingFlushIncompleteBuf(b *testing.B) {
	benchmarkFlushIncompleteBuf(b, NewBuffer(512))
}

func BenchmarkRingFlushFullBuf(b *testing.B) {
	benchmarkFlushFullBuf(b, NewBuffer(512))
}
