package network

import (
	"encoding/binary"
	"fmt"
)

type EthHdr struct {
	Data []byte
}

func (eth *EthHdr) New(data []byte, offset int) error {
	pkt := data[offset:]

	if len(pkt) < 14 {
		return fmt.Errorf("not enough data")
	}

	eth.Data = pkt

	return nil
}

func (eth *EthHdr) GetSrcMac() [6]byte {
	var mac [6]byte

	copy(mac[:], eth.Data[0:6])

	return mac
}

func (eth *EthHdr) SetSrcMac(val [6]byte) {
	copy(eth.Data[0:6], val[:])
}

func (eth *EthHdr) GetDstMac() [6]byte {
	var mac [6]byte

	copy(mac[:], eth.Data[6:12])

	return mac
}

func (eth *EthHdr) SetDstMac(val [6]byte) {
	copy(eth.Data[6:12], val[:])
}

func (eth *EthHdr) GetType() uint16 {
	return binary.BigEndian.Uint16(eth.Data[12:14])
}

func (eth *EthHdr) SetType(val uint16) {
	binary.BigEndian.PutUint16(eth.Data[12:14], val)
}
