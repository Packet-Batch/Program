package network

import (
	"encoding/binary"
	"fmt"
)

type UdpHdr struct {
	Data []byte
}

func (udph *UdpHdr) New(data []byte, offset int) error {
	pkt := data[offset:]

	if len(pkt) < 8 {
		return fmt.Errorf("not enough data")
	}

	udph.Data = pkt

	return nil
}

func (udph *UdpHdr) GetSrcPort() uint16 {
	return binary.BigEndian.Uint16(udph.Data[0:2])
}

func (udph *UdpHdr) SetSrcPort(val uint16) {
	binary.BigEndian.PutUint16(udph.Data[0:2], val)
}

func (udph *UdpHdr) GetDstPort() uint16 {
	return binary.BigEndian.Uint16(udph.Data[2:4])
}

func (udph *UdpHdr) SetDstPort(val uint16) {
	binary.BigEndian.PutUint16(udph.Data[2:4], val)
}

func (udph *UdpHdr) GetLength() uint16 {
	return binary.BigEndian.Uint16(udph.Data[4:6])
}

func (udph *UdpHdr) SetLength(val uint16) {
	binary.BigEndian.PutUint16(udph.Data[4:6], val)
}

func (udph *UdpHdr) GetChecksum() uint16 {
	return binary.BigEndian.Uint16(udph.Data[6:8])
}

func (udph *UdpHdr) SetChecksum(val uint16) {
	binary.BigEndian.PutUint16(udph.Data[6:8], val)
}
