// Package utility provides some basic tools.
package utility

import (
	"container/list"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

// LogQue is the interface definition of the queue logger.
type LogQue interface {

	// Log prints a message immediately.
	Log(msg interface{})

	// Fatal prints the current and all previous messages, then exits.
	Fatal(msg interface{})

	// Panic prints the current and all previous messages, then throws a panic.
	Panic(msg interface{})

	// Store stores a message.
	Store(msg interface{}) int

	// LogStorage prints all previous messages.
	LogStorage()
}

type logQue struct {
	msgs       *list.List
	timeLayout string
	logger     *log.Logger
	mutex      sync.Mutex
}

// NewLogQue creates a new logger.
func NewLogQue(out io.Writer, timeLayout string) LogQue {
	Assert(out != nil, "Null I/O writer.")

	return &logQue{
		msgs:       list.New(),
		logger:     log.New(out, "", 0),
		timeLayout: timeLayout,
	}
}

func (que *logQue) Log(msg interface{}) {
	que.logger.Println(que.format(msg))
}

func (que *logQue) Fatal(msg interface{}) {
	que.LogStorage()
	que.logger.Fatalln(que.format(msg))
}

func (que *logQue) Panic(msg interface{}) {
	que.LogStorage()
	que.logger.Panicln(que.format(msg))
}

func (que *logQue) Store(msg interface{}) int {
	que.mutex.Lock()
	defer que.mutex.Unlock()

	que.msgs.PushBack(que.format(msg))
	return que.msgs.Len()
}

func (que *logQue) LogStorage() {
	que.mutex.Lock()
	defer que.mutex.Unlock()

	for e := que.msgs.Front(); e != nil; e = e.Next() {
		que.logger.Println(e.Value)
	}

	que.msgs = que.msgs.Init()
}

func (que *logQue) format(msg interface{}) string {
	if que.timeLayout != "" {
		return fmt.Sprintf("%v: %v", time.Now().Format(que.timeLayout), msg)
	}

	return fmt.Sprintf("%v", msg)
}
