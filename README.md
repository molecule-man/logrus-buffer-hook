# Fixed-size buffer hook for Logrus <img src="http://i.imgur.com/hTeVwmJ.png" width="40" height="40" alt=":walrus:" class="emoji" title=":walrus:" />&nbsp;[![Build Status](https://circleci.com/gh/molecule-man/logrus-buffer-hook/tree/main.svg?style=svg)](https://circleci.com/gh/molecule-man/logrus-buffer-hook/tree/main)

This logrus hook keeps the logs in fixed-size buffer and flushes the buffer when
certain condition is met, e.g. when error level is logged.

## Usage

Here is an example of setup where `Debug` and `Info` logs are stored in the
buffer and as soon as one of `Warning, Error, Fatal, Panic` levels is logged
then the buffer is flushed. Notice that as the buffer is of fixed size, as long
as you log only `Debug` and `Info` the older logs eventually get overwritten by
newer logs.

```go
package main

import (
	"os"

	"github.com/sirupsen/logrus"
	bufferhook "github.com/molecule-man/logrus-buffer-hook.git"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(ioutil.Discard) // discard main output writer, otherwise logs will be written twice

	hook := bufferhook.New(os.Stderr, 10000) // the buffer will be of fixed size of 10 kb
	hook.LogLevels = []log.Level{
		// this log levels are always logged and cause the buffer with Debug and
		// Info levels to be flushed. This behavior is specified by the next
		// statement where we define FlushCondition
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,

		// this log levels are not logged immediately but rather kept in buffer
		// until one of the previously defined levels is logged
		logrus.InfoLevel,
		logrus.DebugLevel,
	},
	// here we actually define that all levels starting from `Warning` should
	// cause buffer to be flushed
	hook.FlushCondition = bufferhook.FlushOnLevel(logrus.WarnLevel)

	logger.AddHook(hook)
}
```

### Prevent unnecessary serialization from happening

In previous example it was shown that logger's default writer has to be
discarded  with `logger.SetOutput(ioutil.Discard)`. However then logrus will
serialize log entry twice. Once by logger (before sending it to io.Discard) and
the second time by the hook itself. To prevent this unnecessary serialization
it's recommended to use `NullFormatter` that does nothing:

```go
// ...
hook.Formatter = logger.Formatter
// ...
logger.Formatter = &bufferhook.NullFormatter{}
}
```
