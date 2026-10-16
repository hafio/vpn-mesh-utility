package proto_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func TestSSERoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := proto.WriteRetry(&buf, 3*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := proto.WriteComment(&buf, "ka"); err != nil {
		t.Fatal(err)
	}
	if err := proto.WriteEvent(&buf, proto.Event{Name: proto.EventConfig, Data: []byte(`{"version":4,"hash":"ab"}`)}); err != nil {
		t.Fatal(err)
	}
	if err := proto.WriteComment(&buf, "ka"); err != nil {
		t.Fatal(err)
	}
	want := "retry: 3000\n\n: ka\n\nevent: config\ndata: {\"version\":4,\"hash\":\"ab\"}\n\n: ka\n\n"
	if buf.String() != want {
		t.Fatalf("wire format:\n got %q\nwant %q", buf.String(), want)
	}

	r := proto.NewReader(&buf)
	ev, err := r.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if ev.Name != proto.EventConfig || string(ev.Data) != `{"version":4,"hash":"ab"}` {
		t.Fatalf("event = %q %q", ev.Name, ev.Data)
	}
	if _, err := r.Next(); err != io.EOF {
		t.Fatalf("Next at end = %v, want io.EOF", err)
	}
}

func TestSSEReaderJoinsDataLinesAndDefaultsName(t *testing.T) {
	r := proto.NewReader(strings.NewReader("data: a\ndata: b\nid: 9\n\n"))
	ev, err := r.Next()
	if err != nil {
		t.Fatal(err)
	}
	if ev.Name != "message" || string(ev.Data) != "a\nb" {
		t.Fatalf("event = %q %q, want message \"a\\nb\"", ev.Name, ev.Data)
	}
}

func TestSSEReaderReportsTruncatedEvent(t *testing.T) {
	r := proto.NewReader(strings.NewReader("event: config\ndata: {\"version\":1"))
	if _, err := r.Next(); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Next = %v, want io.ErrUnexpectedEOF", err)
	}
}

func TestSSEReaderBoundsLineLength(t *testing.T) {
	huge := "data: " + strings.Repeat("x", 128<<10) + "\n\n"
	if _, err := proto.NewReader(strings.NewReader(huge)).Next(); err == nil || err == io.EOF {
		t.Fatalf("Next on an oversized line = %v, want an error", err)
	}
}

func TestSSEWriterRefusesFrameInjection(t *testing.T) {
	var buf bytes.Buffer
	cases := []proto.Event{
		{Name: "", Data: []byte("x")},
		{Name: "config\nevent: evil", Data: []byte("x")},
		{Name: "con:fig", Data: []byte("x")},
		{Name: "config", Data: []byte("a\n\nevent: evil\ndata: b")},
		{Name: "config", Data: []byte("a\rb")},
	}
	for _, ev := range cases {
		if err := proto.WriteEvent(&buf, ev); err == nil {
			t.Errorf("WriteEvent(%q, %q) accepted a frame-breaking event", ev.Name, ev.Data)
		}
	}
	if err := proto.WriteComment(&buf, "a\nb"); err == nil {
		t.Error("WriteComment accepted a newline")
	}
	if buf.Len() != 0 {
		t.Fatalf("refused frames must write nothing, wrote %q", buf.String())
	}
}
