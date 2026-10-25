package wg

import (
	"fmt"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// The Windows adapter's GUID (amendment A34) is fixed per mesh slot
// (amendment A49): its third part is 0x6c00 plus the slot. Windows then
// knows each mesh's adapter again after every restart -- slot 0's is the
// GUID it always had -- and an agent tells the adapters of meshctl's other
// meshes from any other program's. The parts are kept apart from the
// Windows type so every platform's tests pin them.
const (
	adapterData1 = 0x6d657368
	adapterData2 = 0x6374
	adapterData3 = 0x6c00
)

var adapterData4 = [8]byte{0x8a, 0x2f, 0x41, 0x7e, 0x0c, 0x5b, 0x93, 0xd1}

// adapterPart is the third part of the GUID of the adapter for the
// interface name, which must be a slot's (proto.Slot.Interface).
func adapterPart(name string) (uint16, error) {
	s, ok := proto.InterfaceSlot(name)
	if !ok {
		return 0, fmt.Errorf("create the WireGuard adapter %s: not a mesh's interface name (wg0 to wg%d)", name, proto.MaxSlot)
	}
	return adapterData3 + uint16(s), nil
}

// adapterSlot is the slot of the meshctl adapter whose GUID has these
// parts, if it is one.
func adapterSlot(d1 uint32, d2, d3 uint16, d4 [8]byte) (proto.Slot, bool) {
	s := proto.Slot(int(d3) - adapterData3)
	if d1 != adapterData1 || d2 != adapterData2 || d4 != adapterData4 || !s.Valid() {
		return 0, false
	}
	return s, true
}
