package wg

import (
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Each slot's adapter has a GUID of its own, slot 0's the one it always
// had (amendment A49), and the agent reads a meshctl adapter's slot back
// from its GUID and finds none in anyone else's.
func TestAdapterGUID(t *testing.T) {
	if d3, err := adapterPart("wg0"); err != nil || d3 != 0x6c00 {
		t.Fatalf("wg0 = %#x, %v", d3, err)
	}
	for s := proto.Slot(0); s <= proto.MaxSlot; s++ {
		d3, err := adapterPart(s.Interface())
		if err != nil {
			t.Fatal(err)
		}
		if got, ok := adapterSlot(adapterData1, adapterData2, d3, adapterData4); !ok || got != s {
			t.Fatalf("slot %d reads back as %d, %v", s, got, ok)
		}
	}
	if _, err := adapterPart("mce-test0"); err == nil || !strings.Contains(err.Error(), "wg0 to wg63") {
		t.Fatalf("a name no slot has = %v", err)
	}
	other := adapterData4
	other[7]++
	for _, c := range []struct {
		d1     uint32
		d2, d3 uint16
		d4     [8]byte
	}{
		{0x2b5c0e59, adapterData2, adapterData3, adapterData4},
		{adapterData1, 0x6375, adapterData3, adapterData4},
		{adapterData1, adapterData2, adapterData3 + 64, adapterData4},
		{adapterData1, adapterData2, adapterData3 - 1, adapterData4},
		{adapterData1, adapterData2, adapterData3, other},
	} {
		if s, ok := adapterSlot(c.d1, c.d2, c.d3, c.d4); ok {
			t.Errorf("adapterSlot(%#x, %#x, %#x, %x) = %d", c.d1, c.d2, c.d3, c.d4, s)
		}
	}
}
