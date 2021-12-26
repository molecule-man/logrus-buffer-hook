package logrusbufferhook

import (
	"io"

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

	if !hook.FlushCondition(entry, line, hook.buf) {
		_, err = hook.buf.Write(line)
		return err
	}

	_, err = hook.buf.WriteTo(hook.w)
	if err != nil {
		return err
	}

	_, err = hook.w.Write(line)
	return err
}

func (hook *Hook) Levels() []logrus.Level {
	return hook.LogLevels
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
