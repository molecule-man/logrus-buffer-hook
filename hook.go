package logrusbufferhook

import (
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

type FlushCondition func(*logrus.Entry, []byte, *Buffer) bool

func New(w io.Writer, size int) *Hook {
	if size <= 0 {
		size = 4096
	}

	return &Hook{
		w:              w,
		LogLevels:      logrus.AllLevels,
		FlushCondition: FlushOnBufferOverflow,
		buf:            NewBuffer(size),
	}
}

type Hook struct {
	LogLevels      []logrus.Level
	FlushCondition FlushCondition
	Formatter      logrus.Formatter

	mu  sync.Mutex
	w   io.Writer
	buf *Buffer
}

func (hook *Hook) Fire(entry *logrus.Entry) error {
	var line []byte
	var err error

	if hook.Formatter != nil {
		line, err = hook.Formatter.Format(entry)
	} else {
		line, err = entry.Bytes()
	}

	if err != nil {
		return err
	}

	hook.mu.Lock()
	defer hook.mu.Unlock()

	if !hook.FlushCondition(entry, line, hook.buf) {
		_, err = hook.buf.Write(line)
		return err
	}

	if _, err := hook.buf.WriteTo(hook.w); err != nil {
		return err
	}

	_, err = hook.w.Write(line)
	return err
}

func (hook *Hook) Levels() []logrus.Level {
	return hook.LogLevels
}

// Flush forces underlying buffer to flush its content
func (hook *Hook) Flush() error {
	hook.mu.Lock()
	defer hook.mu.Unlock()

	_, err := hook.buf.WriteTo(hook.w)
	return err
}

func FlushOnLevel(level logrus.Level) FlushCondition {
	return func(entry *logrus.Entry, line []byte, buf *Buffer) bool {
		return entry.Level <= level
	}
}

func FlushOnBufferOverflow(entry *logrus.Entry, line []byte, buf *Buffer) bool {
	return len(line) > buf.Available()
}

type NullFormatter struct{}

func (NullFormatter) Format(e *logrus.Entry) ([]byte, error) { return []byte{}, nil }
