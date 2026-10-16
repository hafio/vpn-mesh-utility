package logbuf

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestRingKeepsTheLastLines(t *testing.T) {
	r := New(3)
	if got := r.Lines(); len(got) != 0 {
		t.Fatalf("a new ring holds %q", got)
	}
	_, _ = io.WriteString(r, "a\nb\n")
	if got := r.Lines(); !slices.Equal(got, []string{"a", "b"}) {
		t.Fatalf("Lines = %q", got)
	}
	_, _ = io.WriteString(r, "c\nd\ne\n")
	if got := r.Lines(); !slices.Equal(got, []string{"c", "d", "e"}) {
		t.Fatalf("after wrapping, Lines = %q", got)
	}
	// A line arrives whole or not at all.
	if n, err := io.WriteString(r, "f"); n != 1 || err != nil || !slices.Equal(r.Lines(), []string{"c", "d", "e"}) {
		t.Fatalf("a partial line: %d, %v, %q", n, err, r.Lines())
	}
	_, _ = io.WriteString(r, "g\n")
	if got := r.Lines(); !slices.Equal(got, []string{"d", "e", "fg"}) {
		t.Fatalf("after the rest of the line, Lines = %q", got)
	}

	_, _ = io.WriteString(r, strings.Repeat("x", 3*maxLine)+"\n")
	if got := r.Lines()[2]; len(got) != maxLine || !strings.HasSuffix(got, "x...") {
		t.Fatalf("a long line is kept as %d bytes ending %q", len(got), got[len(got)-4:])
	}
	if got := New(0); len(got.lines) != 1 {
		t.Fatalf("New(0) keeps %d lines", len(got.lines))
	}
}

func next(t *testing.T, ch <-chan string) (string, bool) {
	t.Helper()
	select {
	case l, ok := <-ch:
		return l, ok
	case <-time.After(5 * time.Second):
		t.Fatal("no line from the follower")
		return "", false
	}
}

// A follower gets what the ring holds, then every new line, until it leaves.
func TestFollow(t *testing.T) {
	r := New(10)
	_, _ = io.WriteString(r, "before\n")
	ctx, cancel := context.WithCancel(context.Background())
	held, ch := r.Follow(ctx)
	if !slices.Equal(held, []string{"before"}) {
		t.Fatalf("held = %q", held)
	}
	_, _ = io.WriteString(r, "after\n")
	if l, ok := next(t, ch); !ok || l != "after" {
		t.Fatalf("followed %q, %v", l, ok)
	}
	cancel()
	if _, ok := next(t, ch); ok {
		t.Fatal("the channel stayed open after the follower left")
	}
	_, _ = io.WriteString(r, "later\n") // nobody follows; must not block
}

// A follower that stops reading is dropped, and logging never waits for it.
func TestSlowFollowerIsDropped(t *testing.T) {
	r := New(10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, ch := r.Follow(ctx)
	done := make(chan struct{})
	go func() {
		for i := range followBuffer + 1 {
			fmt.Fprintf(r, "line %d\n", i)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("a slow follower held up the log")
	}
	n := 0
	for range ch {
		n++
	}
	if n != followBuffer {
		t.Fatalf("the dropped follower got %d lines, want the %d it had room for", n, followBuffer)
	}
	if got := r.Lines(); len(got) != 10 || got[9] != fmt.Sprintf("line %d", followBuffer) {
		t.Fatalf("the ring after the drop: %q", got)
	}
}
