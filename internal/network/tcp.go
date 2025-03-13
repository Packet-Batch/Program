package network

import (
	"encoding/binary"
	"fmt"
)

type TcpHdr struct {
	Data []byte
}

const (
	TCP_FIN uint16 = 1 << 0
	TCP_SYN uint16 = 1 << 1
	TCP_RST uint16 = 1 << 2
	TCP_PSH uint16 = 1 << 3
	TCP_ACK uint16 = 1 << 4
	TCP_URG uint16 = 1 << 5
	TCP_ECE uint16 = 1 << 6
	TCP_CWR uint16 = 1 << 7
)

func (tcph *TcpHdr) New(pkt []byte, offset int) error {
	tcph.Data = pkt[offset:]

	if len(tcph.Data) < 20 {
		return fmt.Errorf("not enough data")
	}

	return nil
}

func (tcph *TcpHdr) GetSrcPort() uint16 {
	return binary.BigEndian.Uint16(tcph.Data[0:2])
}

func (tcph *TcpHdr) SetSrcPort(val uint16) {
	binary.BigEndian.PutUint16(tcph.Data[0:2], val)
}

func (tcph *TcpHdr) GetDstPort() uint16 {
	return binary.BigEndian.Uint16(tcph.Data[2:4])
}

func (tcph *TcpHdr) SetDstPort(val uint16) {
	binary.BigEndian.PutUint16(tcph.Data[2:4], val)
}

func (tcph *TcpHdr) GetSeqNum() uint32 {
	return binary.BigEndian.Uint32(tcph.Data[4:8])
}

func (tcph *TcpHdr) SetSeqNum(val uint32) {
	binary.BigEndian.PutUint32(tcph.Data[4:8], val)
}

func (tcph *TcpHdr) GetAckNum() uint32 {
	return binary.BigEndian.Uint32(tcph.Data[8:12])
}

func (tcph *TcpHdr) SetAckNum(val uint32) {
	binary.BigEndian.PutUint32(tcph.Data[8:12], val)
}

func (tcph *TcpHdr) GetHdrLen() uint8 {
	return 0
}

func (tcph *TcpHdr) SetHdrLen(val uint8) {

}

func (tcph *TcpHdr) GetFlags() uint16 {
	return binary.BigEndian.Uint16(tcph.Data[13:15])
}

func (tcph *TcpHdr) SetFlags(val uint16) {
	binary.BigEndian.PutUint16(tcph.Data[13:15], val)
}

func (tcph *TcpHdr) GetWindow() uint16 {
	return binary.BigEndian.Uint16(tcph.Data[16:18])
}

func (tcph *TcpHdr) SetWindow(val uint16) {
	binary.BigEndian.PutUint16(tcph.Data[16:18], val)
}

func (tcph *TcpHdr) GetCheck() uint16 {
	return binary.BigEndian.Uint16(tcph.Data[18:20])
}

func (tcph *TcpHdr) SetCheck(val uint16) {
	binary.BigEndian.PutUint16(tcph.Data[18:20], val)
}

func (tcph *TcpHdr) GetUrgPtr() uint16 {
	return binary.BigEndian.Uint16(tcph.Data[20:22])
}

func (tcph *TcpHdr) SetUrgPtr(val uint16) {
	binary.BigEndian.PutUint16(tcph.Data[20:22], val)
}
