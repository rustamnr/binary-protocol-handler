package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
)

const HeaderSize = 24

type Header struct {
	Type     uint32
	ID       uint32
	Response uint32
	Param    uint32
	DataLen  uint32
	Checksum uint32
}

func ReadHeader(r io.Reader) (*Header, error) {
	buf := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	log.Printf("Read header: %x", buf)

	h := &Header{
		Type:     binary.BigEndian.Uint32(buf[0:4]),
		ID:       binary.BigEndian.Uint32(buf[4:8]),
		Response: binary.BigEndian.Uint32(buf[8:12]),
		Param:    binary.BigEndian.Uint32(buf[12:16]),
		DataLen:  binary.BigEndian.Uint32(buf[16:20]),
		Checksum: binary.BigEndian.Uint32(buf[20:24]),
	}

	expectedChecksum := crc32.ChecksumIEEE(buf[:20])
	if expectedChecksum != h.Checksum {
		return nil, fmt.Errorf("checksum error: expected %08x, got %08x", expectedChecksum, h.Checksum)
	}

	return h, nil
}

func WriteHeader(w io.Writer, h Header) error {
	buf := new(bytes.Buffer)
	fields := []uint32{h.Type, h.ID, h.Response, h.Param, h.DataLen}
	for _, v := range fields {
		if err := binary.Write(buf, binary.BigEndian, v); err != nil {
			return err
		}
	}
	checksum := crc32.ChecksumIEEE(buf.Bytes())
	binary.Write(buf, binary.BigEndian, checksum)
	_, err := w.Write(buf.Bytes())
	return err
}
