package gittransport

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

var (
	// ErrInvalidPktLine is returned when a pkt-line has invalid hex or format.
	ErrInvalidPktLine = errors.New("gittransport: invalid pkt-line")
	// ErrLineTooLong is returned when pkt-line length exceeds Git protocol maximum 65520 bytes.
	ErrLineTooLong = errors.New("gittransport: pkt-line length exceeds 65520")
)

// PktType defines the special Git pkt-line types.
type PktType int

const (
	PktData PktType = iota
	PktFlush
	PktDelim
	PktResponseEnd
)

// Packet represents a parsed Git pkt-line.
type Packet struct {
	Type    PktType
	Payload []byte
}

// ReadPacket parses one Git pkt-line from r.
func ReadPacket(r io.Reader) (Packet, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return Packet{}, err
	}

	lengthStr := string(lenBuf[:])
	switch lengthStr {
	case "0000":
		return Packet{Type: PktFlush}, nil
	case "0001":
		return Packet{Type: PktDelim}, nil
	case "0002":
		return Packet{Type: PktResponseEnd}, nil
	}

	n, err := strconv.ParseUint(lengthStr, 16, 16)
	if err != nil {
		return Packet{}, fmt.Errorf("%w: bad hex length %q", ErrInvalidPktLine, lengthStr)
	}

	if n < 4 {
		return Packet{}, fmt.Errorf("%w: length %d less than header", ErrInvalidPktLine, n)
	}
	if n > 65520 {
		return Packet{}, ErrLineTooLong
	}

	payloadLen := int(n - 4)
	if payloadLen == 0 {
		return Packet{Type: PktData, Payload: nil}, nil
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return Packet{}, fmt.Errorf("%w: incomplete payload: %v", ErrInvalidPktLine, err)
	}

	return Packet{Type: PktData, Payload: payload}, nil
}

// ParsePacketLine parses an in-memory buffer into a Packet.
func ParsePacketLine(data []byte) (Packet, error) {
	return ReadPacket(bytes.NewReader(data))
}
