package protocol_test

import (
	"bytes"
	"testing"

	"github.com/rustamnr/binary-protocol-handler/internal/protocol"
)

func TestReadWriteHeader(t *testing.T) {
	original := protocol.Header{
		Type:     1,
		ID:       42,
		Response: 200,
		Param:    777,
		DataLen:  11,
	}

	buf := &bytes.Buffer{}
	err := protocol.WriteHeader(buf, original)
	if err != nil {
		t.Fatalf("WriteHeader failed: %v", err)
	}

	read, err := protocol.ReadHeader(buf)
	if err != nil {
		t.Fatalf("ReadHeader failed: %v", err)
	}

	// Обновляем оригинальный Checksum, чтобы сравнение было честным
	original.Checksum = read.Checksum

	if *read != original {
		t.Errorf("Header mismatch.\nExpected: %+v\nGot:      %+v", original, *read)
	}
}

func TestReadHeader_InvalidChecksum(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, protocol.HeaderSize))
	_, err := protocol.ReadHeader(buf)
	if err == nil {
		t.Errorf("Ожидалась ошибка контрольной суммы, но её не было")
	}
}
