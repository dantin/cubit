package app

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/dantin/cubit/version"
	"github.com/stretchr/testify/require"
)

type writerBuffer struct {
	mu      sync.RWMutex
	buf     *bytes.Buffer
	closeCh chan bool
}

func newWriterBuffer() *writerBuffer {
	return &writerBuffer{buf: bytes.NewBuffer(nil), closeCh: make(chan bool)}
}

func (wb *writerBuffer) Write(p []byte) (int, error) {
	wb.mu.Lock()
	defer wb.mu.Unlock()
	return wb.buf.Write(p)
}

func (wb *writerBuffer) String() string {
	wb.mu.RLock()
	defer wb.mu.RUnlock()
	return wb.buf.String()
}

func TestApp_EmptyArgs(t *testing.T) {
	require.NotNil(t, New(nil, nil))
}

func TestApp_ShowUsage(t *testing.T) {
	w := newWriterBuffer()
	err := New(w, []string{"./cubit", "-h"}).Run()
	require.Nil(t, err)
	require.Equal(t, expectedUsageString(), w.String())
}

func TestApp_PrintVersion(t *testing.T) {
	w := newWriterBuffer()
	args := []string{"./cubit", "--version"}
	err := New(w, args).Run()
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf("cubit version: %v\n", version.ApplicationVersion), w.String())
}

func TestApplication_Run(t *testing.T) {
	w := newWriterBuffer()
	args := []string{"./cubit", "--config=../data/basic.yml"}
	ap := New(w, args)
	go func() {
		time.Sleep(time.Millisecond * 1500) // wait until initialized
		ap.waitStopCh <- syscall.SIGTERM
	}()
	ap.shutDownWaitSecs = time.Duration(2) * time.Second // wait only two seconds
	err := ap.Run()
	require.Nil(t, err)

	os.RemoveAll(".cert/")

	// make sure pid and log files had been created
	_, err = os.Stat("ut.cubit.pid")
	require.False(t, os.IsNotExist(err))
	os.Remove("ut.cubit.pid")

	_, err = os.Stat("ut.cubit.log")
	require.False(t, os.IsNotExist(err))
	os.Remove("ut.cubit.log")
}

func expectedUsageString() string {
	var r string
	for i := range logoStr {
		r += fmt.Sprintf("%s\n", logoStr[i])
	}
	r += fmt.Sprintf("%s\n", usageStr)
	return r
}
