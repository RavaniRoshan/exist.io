package gittransport

import (
	"bytes"
	"testing"
)

func FuzzPacketLineParser(f *testing.F) {
	// Seed corpus
	f.Add([]byte("0000"))
	f.Add([]byte("0001"))
	f.Add([]byte("0002"))
	f.Add([]byte("0004"))
	f.Add([]byte("000ahello\n"))
	f.Add([]byte("0014command=ls-refs\n"))
	f.Add([]byte("ffff"))
	f.Add([]byte(""))
	f.Add([]byte("000"))
	f.Add([]byte("zzzz"))

	f.Fuzz(func(t *testing.T, data []byte) {
		pkt, err := ReadPacket(bytes.NewReader(data))
		if err != nil {
			// Parsing rejected invalid input without panicking
			return
		}

		// Valid packet sanity checks
		switch pkt.Type {
		case PktFlush, PktDelim, PktResponseEnd:
			if len(pkt.Payload) != 0 {
				t.Fatalf("expected empty payload for special pkt type %v, got %d bytes", pkt.Type, len(pkt.Payload))
			}
		case PktData:
			if len(pkt.Payload) > 65520-4 {
				t.Fatalf("payload length %d exceeds max git pkt size", len(pkt.Payload))
			}
		}
	})
}
