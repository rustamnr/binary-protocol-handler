// internal/worker/worker_test.go
package worker_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/rustamnr/binary-protocol-handler/internal/processor"
	"github.com/rustamnr/binary-protocol-handler/internal/protocol"
	"github.com/rustamnr/binary-protocol-handler/internal/worker"
)

type mockConn struct {
	bytes.Buffer
}

func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestWorkerProcessMessage(t *testing.T) {
	conn := &mockConn{}
	header := protocol.Header{
		Type:     1,
		ID:       123,
		Response: 0,
		Param:    0,
		DataLen:  4,
	}
	task := worker.MessageTask{
		Conn:   conn,
		Header: header,
		Data:   []byte{1, 2, 3, 4}, // sum = 10
	}

	tasks := make(chan worker.MessageTask, 1)
	tasks <- task
	close(tasks)

	p := &processor.DummyProcessor{}
	worker.Start(tasks, p)

	// Проверим, что в буфере conn есть записанный заголовок ответа
	resp, err := protocol.ReadHeader(conn)
	if err != nil {
		t.Fatalf("Не удалось прочитать заголовок ответа: %v", err)
	}

	if resp.ID != 123 || resp.Response != 10 {
		t.Errorf("Неверный ответ: %+v", resp)
	}
}
