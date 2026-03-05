package golog

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func newTestLogger(level int) (*MyLogger, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	ml := New("", level, 0, &stdout, &stderr)
	return ml, &stdout, &stderr
}

func TestLevelFiltering(t *testing.T) {
	ml, stdout, stderr := newTestLogger(LevelWarning)

	ml.Debug("debug msg")
	ml.Info("info msg")
	ml.Notice("notice msg")
	ml.Warning("warning msg")
	ml.Error("error msg")
	ml.Critical("critical msg")

	out := stdout.String()
	errOut := stderr.String()

	if strings.Contains(out, "[Debug]") {
		t.Error("Debug should not appear at LevelWarning")
	}
	if strings.Contains(out, "[Info]") {
		t.Error("Info should not appear at LevelWarning")
	}
	if strings.Contains(out, "[Notice]") {
		t.Error("Notice should not appear at LevelWarning")
	}
	if !strings.Contains(out, "[Warning]") {
		t.Error("Warning should appear at LevelWarning")
	}
	if !strings.Contains(errOut, "[Error]") {
		t.Error("Error should appear at LevelWarning")
	}
	if !strings.Contains(errOut, "[Critical]") {
		t.Error("Critical should appear at LevelWarning")
	}
}

func TestTraceOnlyAtLevelTrace(t *testing.T) {
	ml, stdout, _ := newTestLogger(LevelDebug)

	ml.Trace("trace msg")
	ml.Traceln("trace msg")
	ml.Tracef("trace %s", "msg")

	if stdout.Len() != 0 {
		t.Error("Trace messages should not appear at LevelDebug")
	}

	ml.SetLogLevel(LevelTrace)
	ml.Trace("trace msg")
	if !strings.Contains(stdout.String(), "[Trace]") {
		t.Error("Trace messages should appear at LevelTrace")
	}
}

func TestPrintUsesStdout(t *testing.T) {
	ml, stdout, stderr := newTestLogger(LevelNotice)

	ml.Print("print msg")
	ml.Println("println msg")
	ml.Printf("printf %s", "msg")

	if stdout.Len() == 0 {
		t.Error("Print methods should write to stdout")
	}
	if strings.Contains(stderr.String(), "[Notice]") {
		t.Error("Print methods should not write to stderr")
	}
}

func TestPanicProducesString(t *testing.T) {
	ml, _, _ := newTestLogger(LevelCritical)

	t.Run("Panic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
			if _, ok := r.(string); !ok {
				t.Errorf("expected string panic value, got %T", r)
			}
		}()
		ml.Panic("test panic")
	})

	t.Run("Panicf", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
			s, ok := r.(string)
			if !ok {
				t.Errorf("expected string panic value, got %T", r)
			}
			expected := fmt.Sprintf("panic %d", 42)
			if s != expected {
				t.Errorf("expected %q, got %q", expected, s)
			}
		}()
		ml.Panicf("panic %d", 42)
	})
}

func TestConcurrentGet(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := fmt.Sprintf("test/pkg/%d", n%10)
			logger := Get(name)
			if logger == nil {
				t.Errorf("Get(%q) returned nil", name)
			}
		}(i)
	}
	wg.Wait()
}

func TestLevelNoneDisablesAll(t *testing.T) {
	ml, stdout, stderr := newTestLogger(LevelNone)

	ml.Critical("crit")
	ml.Error("err")
	ml.Warning("warn")
	ml.Notice("notice")
	ml.Info("info")
	ml.Debug("debug")
	ml.Trace("trace")

	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Error("LevelNone should suppress all output")
	}
}
