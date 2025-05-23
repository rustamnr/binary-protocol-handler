package processor_test

import (
	"bytes"
	"testing"

	"github.com/rustamnr/binary-protocol-handler/internal/processor"
)

func TestDummyProcessor(t *testing.T) {
	p := &processor.DummyProcessor{}
	data := []byte{1, 2, 3, 4} // sum = 10

	result := p.ProcessMessage(0, bytes.NewReader(data))
	if result != 10 {
		t.Errorf("expected 10, got %d", result)
	}
}
