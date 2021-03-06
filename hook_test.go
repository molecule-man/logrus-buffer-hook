package logrusbufferhook_test

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/quick"

	logrusbufferhook "github.com/Molecule-man/logrus-buffer-hook"
	"github.com/sirupsen/logrus"
)

func TestFlushOnBufferOverflow(t *testing.T) {
	expectedBuf := &bytes.Buffer{}
	testedBuf := &bytes.Buffer{}

	logger := logrus.New()
	logger.SetFormatter(&testFormatter{})
	logger.SetOutput(expectedBuf)

	logsNum, flushesNum := 0, 0

	logger.AddHook(logrusbufferhook.New(testedBuf, 1024))

	if err := quick.Check(func(msg string) bool {
		logsNum++

		if len(msg) > 511 {
			msg = msg[:511]
		}

		bytesFlushedSoFar := testedBuf.Len()
		logger.Info(msg)

		if testedBuf.Len() > bytesFlushedSoFar {
			flushesNum++
			return expectedBuf.String() == testedBuf.String()
		}

		return expectedBuf.Len() > testedBuf.Len()
	}, nil); err != nil {
		t.Error(err)
	}

	if logsNum == 0 || flushesNum == 0 || flushesNum > logsNum/2 {
		t.Errorf("it's not statistically possible to have %d logs and %d flushes", logsNum, flushesNum)
	}
}

func TestConcurrentLogging(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := logrus.New()
	logger.SetFormatter(&logrusbufferhook.NullFormatter{})
	logger.SetOutput(io.Discard)

	hook := logrusbufferhook.New(buf, 30)
	hook.Formatter = &testFormatter{}
	logger.AddHook(hook)

	experimentsNum := 100

	wg := sync.WaitGroup{}
	wg.Add(experimentsNum)

	for i := 0; i < experimentsNum; i++ {
		go func(i int) {
			logger.Infof("log %d", i)
			wg.Done()
		}(i)
	}

	wg.Wait()
	hook.Flush()

	writtenLogs := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if len(writtenLogs) != experimentsNum {
		t.Fatalf("expected to have %d logs, got: %d", experimentsNum, len(writtenLogs))
	}
}

func TestFlushOnError(t *testing.T) {
	testedBuf := &bytes.Buffer{}

	logger := logrus.New()
	logger.SetFormatter(&testFormatter{})
	logger.SetOutput(io.Discard)
	logger.SetLevel(logrus.DebugLevel)

	hook := logrusbufferhook.New(testedBuf, 1024)
	hook.FlushCondition = logrusbufferhook.FlushOnLevel(logrus.ErrorLevel)
	logger.AddHook(hook)

	logger.Error("error 1")
	logger.Error("error 2")
	logger.Error("error 3")

	expected := []string{
		"error 1",
		"error 2",
		"error 3",
	}
	got := strings.Split(strings.TrimSpace(testedBuf.String()), "\n")
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected: %#v, got: %#v", expected, got)
	}

	testedBuf.Reset()

	logger.Info("info 1")
	logger.Debug("debug 2")
	logger.Warn("warning 3")

	if testedBuf.Len() > 0 {
		t.Errorf("expected no logs to be flushed but got %#v", testedBuf.String())
	}

	logger.Error("error 4")

	expected = []string{
		"info 1",
		"debug 2",
		"warning 3",
		"error 4",
	}
	got = strings.Split(strings.TrimSpace(testedBuf.String()), "\n")
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected: %#v, got: %#v", expected, got)
	}
}

type testFormatter struct{}

func (f *testFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message + "\n"), nil
}

func BenchmarkLoggerNoFormatter(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	hook := logrusbufferhook.New(io.Discard, 1024)
	hook.FlushCondition = neverFlush
	logger.AddHook(hook)

	for i := 0; i < b.N; i++ {
		logger.Info("test log")
	}
}

func BenchmarkLoggerNullFormatter(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	hook := logrusbufferhook.New(io.Discard, 1024)
	hook.FlushCondition = neverFlush
	hook.Formatter = logger.Formatter
	logger.AddHook(hook)
	logger.Formatter = &logrusbufferhook.NullFormatter{}

	for i := 0; i < b.N; i++ {
		logger.Info("test log")
	}
}

func neverFlush(_ *logrus.Entry, _ []byte, _ *logrusbufferhook.Buffer) bool { return false }
