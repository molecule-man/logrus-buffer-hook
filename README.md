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
	// discard main output writer, otherwise logs will be written twice
	logger.SetOutput(ioutil.Discard)

	// the buffer will be of fixed size of 10 kb
	hook := bufferhook.New(os.Stderr, 10000)
	// here we  define that all levels starting from `Warning` (so also `Error`,
	// `Fatal` and `Panic`) should cause buffer (containing `Info` and `Debug`
	// logs) to be flushed
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
```
