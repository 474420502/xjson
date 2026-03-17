package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestMainRuns(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	main()
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if buf.Len() == 0 {
		t.Fatal("expected example main to print output")
	}
}