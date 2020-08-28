package log

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type writeBuffer struct {
	mu      sync.RWMutex
	buf     *bytes.Buffer
	closeCh chan bool
}

func newWriterBuffer() *writeBuffer {
	return &writeBuffer{buf: bytes.NewBuffer(nil), closeCh: make(chan bool)}
}

func (wb *writeBuffer) Write(p []byte) (int, error) {
	wb.mu.Lock()
	defer wb.mu.Unlock()
	return wb.buf.Write(p)
}

func (wb *writeBuffer) String() string {
	wb.mu.RLock()
	defer wb.mu.RUnlock()
	return wb.buf.String()
}

func (wb *writeBuffer) Reset() {
	wb.mu.Lock()
	defer wb.mu.Unlock()
	wb.buf.Reset()
}

func (wb *writeBuffer) Close() error {
	wb.closeCh <- true
	return nil
}

func (wb *writeBuffer) Done() <-chan bool {
	return wb.closeCh
}

func TestDebugLog(t *testing.T) {
	bw, _, tearDown := setupTest("debug")
	defer tearDown()

	Debugf("test debug log!")
	time.Sleep(time.Millisecond * 250)

	l := bw.String()

	require.True(t, strings.Contains(l, "[DBG]"))
	require.True(t, strings.Contains(l, "test debug log!"))
}

func TestInfoLog(t *testing.T) {
	bw, _, tearDown := setupTest("info")
	defer tearDown()

	Infof("test info log!")
	time.Sleep(time.Millisecond * 250)

	l := bw.String()

	require.True(t, strings.Contains(l, "[INF]"))
	require.True(t, strings.Contains(l, "test info log!"))
}

func TestWarningLog(t *testing.T) {
	bw, _, tearDown := setupTest("warning")
	defer tearDown()

	Warnf("test warning log!")
	time.Sleep(time.Millisecond * 250)

	l := bw.String()

	require.True(t, strings.Contains(l, "[WRN]"))
	require.True(t, strings.Contains(l, "test warning log!"))
}

func TestErrorLog(t *testing.T) {
	bw, _, tearDown := setupTest("error")
	defer tearDown()

	Errorf("test error log!")
	time.Sleep(time.Millisecond * 250)

	l := bw.String()

	require.True(t, strings.Contains(l, "[ERR]"))
	require.True(t, strings.Contains(l, "test error log!"))

	bw.Reset()

	Error(errors.New("some error string"))
	time.Sleep(time.Millisecond * 250)

	l = bw.String()
	require.True(t, strings.Contains(l, "some error string"))
}

func TestFatalLog(t *testing.T) {
	var exited bool
	exitHandler = func() {
		exited = true
	}

	bw, _, tearDown := setupTest("fatal")
	defer tearDown()

	Fatalf("test fatal log!")
	time.Sleep(time.Millisecond * 250)

	require.True(t, exited)

	l := bw.String()
	require.True(t, strings.Contains(l, "[FTL]"))
	require.True(t, strings.Contains(l, "test fatal log!"))

	bw.Reset()
	exited = false

	Fatal(errors.New("some error string"))
	time.Sleep(time.Millisecond * 250)

	l = bw.String()
	require.True(t, strings.Contains(l, "some error string"))
}

func TestLogFile(t *testing.T) {
	bw, lf, tearDown := setupTest("debug")

	Debugf("test debug log!")
	time.Sleep(time.Millisecond * 250)

	require.Equal(t, bw.String(), lf.String())

	// make sure file is closed
	tearDown()

	select {
	case <-lf.Done():
		require.True(t, true)
	case <-time.After(time.Second):
		require.FailNow(t, "log file has not been closed")
	}
}

func setupTest(level string) (*writeBuffer, *writeBuffer, func()) {
	output := newWriterBuffer()
	logFile := newWriterBuffer()
	l, _ := New(level, output, logFile)
	Set(l)
	return output, logFile, func() { Unset() }
}
