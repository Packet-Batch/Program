package network

import (
	"encoding/binary"
	"fmt"
)

type IpHdr struct {
	Data []byte
}

func (iph *IpHdr) New(pkt []byte, offset int) error {
	iph.Data = pkt[offset:]

	if len(iph.Data) < 20 {
		return fmt.Errorf("not enough data")
	}

	return nil
}

func (iph *IpHdr) GetVersion() uint8 {
	return iph.Data[0] >> 4
}

func (iph *IpHdr) SetVersion(val uint8) {
	iph.Data[0] = (iph.Data[0] & 0x0F) | (val << 4)
}

func (iph *IpHdr) GetIhl() uint8 {
	return iph.Data[0] & 0x0F
}

func (iph *IpHdr) SetIhl(val uint8) {
	iph.Data[0] = (iph.Data[0] & 0xF0) | (val & 0x0F)
}

func (iph *IpHdr) GetTos() uint8 {
	return iph.Data[1]
}

func (iph *IpHdr) SetTos(val uint8) {
	iph.Data[1] = val
}

func (iph *IpHdr) GetLength() uint16 {
	return binary.BigEndian.Uint16(iph.Data[2:4])
}

func (iph *IpHdr) SetLength(val uint16) {
	binary.BigEndian.PutUint16(iph.Data[2:4], val)
}

func (iph *IpHdr) GetId() uint16 {
	return binary.BigEndian.Uint16(iph.Data[4:6])
}

func (iph *IpHdr) SetId(val uint16) {
	binary.BigEndian.PutUint16(iph.Data[4:6], val)
}

func (iph *IpHdr) GetFlags() uint8 {
	return iph.Data[6] >> 5
}

func (iph *IpHdr) SetFlags(val uint8) {
	val &= 0x07

	iph.Data[6] = (iph.Data[6] & 0x1F) | (val << 5)
}

func (iph *IpHdr) GetFragOffset() uint16 {
	return binary.BigEndian.Uint16(iph.Data[6:8]) & 0x1FFF
}

func (iph *IpHdr) SetFragOffset(val uint16) {
	val &= 0x1FFF

	iph.Data[6] = (iph.Data[6] & 0xE0) | uint8(val>>8)
	iph.Data[7] = uint8(val & 0xFF)
}

func (iph *IpHdr) GetTtl() uint8 {
	return iph.Data[8]
}

func (iph *IpHdr) SetTtl(val uint8) {
	iph.Data[8] = val
}

func (iph *IpHdr) GetProtocol() uint8 {
	return iph.Data[9]
}

func (iph *IpHdr) SetProtocol(val uint8) {
	iph.Data[9] = val
}

func (iph *IpHdr) CheckChecksum() uint16 {
	return binary.BigEndian.Uint16(iph.Data[10:12])
}

func (iph *IpHdr) SetChecksum(val uint16) {
	binary.BigEndian.PutUint16(iph.Data[10:12], val)
}

func (iph *IpHdr) GetSrcAddr() uint32 {
	return binary.BigEndian.Uint32(iph.Data[12:16])
}

func (iph *IpHdr) SetSrcAddr(val uint32) {
	binary.BigEndian.PutUint32(iph.Data[12:16], val)
}

func (iph *IpHdr) GetDstAddr() uint32 {
	return binary.BigEndian.Uint32(iph.Data[16:20])
}

func (iph *IpHdr) SetDstAddr(val uint32) {
	binary.BigEndian.PutUint32(iph.Data[16:20], val)
}
