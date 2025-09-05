package client

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"password-manager/internal/common/protocol"
)

// MockServer для тестирования клиента
type MockServer struct {
	listener net.Listener
	handler  func(net.Conn)
}

func NewMockServer(handler func(net.Conn)) *MockServer {
	return &MockServer{handler: handler}
}

func (m *MockServer) Start() error {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}
	m.listener = listener

	go func() {
		for {
			conn, err := m.listener.Accept()
			if err != nil {
				return
			}
			go m.handler(conn)
		}
	}()

	return nil
}

func (m *MockServer) Addr() string {
	return m.listener.Addr().String()
}

func (m *MockServer) Stop() {
	if m.listener != nil {
		m.listener.Close()
	}
}

func TestClientConnect(t *testing.T) {
	server := NewMockServer(func(conn net.Conn) {
		conn.Close()
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	// Извлекаем порт из адреса сервера
	client := NewClient("localhost", 0)

	// Заменяем соединение на mock
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect to mock server: %v", err)
	}
	client.conn = conn

	defer client.Close()

	if client.conn == nil {
		t.Error("Client connection should not be nil")
	}
}

func TestClientRegister(t *testing.T) {
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		// Читаем запрос
		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		header, _ := protocol.DeserializeHeader(headerBuf)
		payload := make([]byte, header.Length)
		conn.Read(payload)

		// Отправляем успешный ответ
		resp := protocol.RegisterResponse{
			Success: true,
			Message: "User registered",
		}
		respData, _ := protocol.SerializeRegisterResponse(resp)
		message := protocol.SerializeMessage(protocol.MsgTypeRegisterResponse, 1, respData)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	defer client.Close()

	err = client.Register("testuser", "testpass")
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}
}

func TestClientLogin(t *testing.T) {
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		header, _ := protocol.DeserializeHeader(headerBuf)
		payload := make([]byte, header.Length)
		conn.Read(payload)

		resp := protocol.AuthResponse{
			Success: true,
			Token:   "test-token",
		}
		respData, _ := protocol.SerializeAuthResponse(resp)
		message := protocol.SerializeMessage(protocol.MsgTypeAuthResponse, 1, respData)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	defer client.Close()

	err = client.Login("testuser", "testpass")
	if err != nil {
		t.Errorf("Login failed: %v", err)
	}

	if !client.IsAuthenticated() {
		t.Error("Client should be authenticated after successful login")
	}

	if client.GetUsername() != "testuser" {
		t.Errorf("Username mismatch. Got: %s, Expected: testuser", client.GetUsername())
	}
}

func TestClientSyncData(t *testing.T) {
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		header, _ := protocol.DeserializeHeader(headerBuf)
		payload := make([]byte, header.Length)
		conn.Read(payload)

		items := []protocol.DataItem{
			{
				ID:   "1",
				Type: protocol.DataTypeText,
				Name: "Test Item",
			},
		}
		resp := protocol.SyncResponse{Items: items}
		respData, _ := protocol.SerializeSyncResponse(resp)
		message := protocol.SerializeMessage(protocol.MsgTypeSyncResponse, 1, respData)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	client.username = "testuser" // имитируем аутентификацию
	client.token = "test-token"
	defer client.Close()

	items, err := client.SyncData(time.Time{})
	if err != nil {
		t.Errorf("SyncData failed: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	}

	if items[0].Name != "Test Item" {
		t.Errorf("Item name mismatch. Got: %s, Expected: Test Item", items[0].Name)
	}
}

func TestClientSaveData(t *testing.T) {
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		header, _ := protocol.DeserializeHeader(headerBuf)
		payload := make([]byte, header.Length)
		conn.Read(payload)

		// Проверяем, что запрос корректный
		req, err := protocol.DeserializeSaveDataRequest(payload)
		if err != nil {
			t.Errorf("Failed to parse save data request: %v", err)
		}

		if req.Item.Name != "Test Item" {
			t.Errorf("Item name mismatch in request: %s", req.Item.Name)
		}

		resp := protocol.SaveDataResponse{
			Success: true,
			Message: "Data saved",
		}
		respData, _ := protocol.SerializeSaveDataResponse(resp)
		message := protocol.SerializeMessage(protocol.MsgTypeSaveDataResponse, 1, respData)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	client.username = "testuser"
	client.token = "test-token"
	defer client.Close()

	item := protocol.NewDataItem{
		Type: protocol.DataTypeText,
		Name: "Test Item",
		Data: []byte("test data"),
		Metadata: map[string]string{
			"test": "value",
		},
	}

	err = client.SaveData(item)
	if err != nil {
		t.Errorf("SaveData failed: %v", err)
	}
}

func TestClientNotAuthenticated(t *testing.T) {
	client := NewClient("localhost", 8080)

	// Все методы должны возвращать ошибку без аутентификации
	_, err := client.SyncData(time.Time{})
	if err == nil {
		t.Error("SyncData should fail when not authenticated")
	}

	err = client.SaveData(protocol.NewDataItem{})
	if err == nil {
		t.Error("SaveData should fail when not authenticated")
	}

	_, err = client.GetData("test-id")
	if err == nil {
		t.Error("GetData should fail when not authenticated")
	}
}

func TestClientErrorResponse(t *testing.T) {
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		// Отправляем ошибку
		errorResp := protocol.ErrorResponse{
			Code:    500,
			Message: "Test error",
		}
		respData, _ := protocol.SerializeErrorResponse(errorResp)
		message := protocol.SerializeMessage(protocol.MsgTypeError, 1, respData)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	defer client.Close()

	err = client.Register("testuser", "testpass")
	if err == nil {
		t.Error("Should fail with error response")
	}

	if err.Error() != "server error: Test error" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestClientUpdateData(t *testing.T) {
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		header, _ := protocol.DeserializeHeader(headerBuf)
		payload := make([]byte, header.Length)
		conn.Read(payload)

		resp := protocol.UpdateDataResponse{
			Success: true,
			Message: "Data updated",
		}
		respData, _ := protocol.SerializeUpdateDataResponse(resp)
		message := protocol.SerializeMessage(protocol.MsgTypeUpdateDataResponse, 1, respData)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	client.username = "testuser"
	client.token = "test-token"
	defer client.Close()

	item := protocol.NewDataItem{
		Type: protocol.DataTypeText,
		Name: "Updated Item",
		Data: []byte("updated data"),
	}

	err = client.UpdateData("test-id", item)
	if err != nil {
		t.Errorf("UpdateData failed: %v", err)
	}
}

func TestClientDeleteData(t *testing.T) {
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		header, _ := protocol.DeserializeHeader(headerBuf)
		payload := make([]byte, header.Length)
		conn.Read(payload)

		req, err := protocol.DeserializeDeleteDataRequest(payload)
		if err != nil {
			t.Errorf("Failed to parse delete request: %v", err)
		}

		if req.ItemID != "test-id" {
			t.Errorf("ItemID mismatch in request: %s", req.ItemID)
		}

		resp := protocol.DeleteDataResponse{
			Success: true,
			Message: "Data deleted",
		}
		respData, _ := protocol.SerializeDeleteDataResponse(resp)
		message := protocol.SerializeMessage(protocol.MsgTypeDeleteDataResponse, 1, respData)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	client.username = "testuser"
	client.token = "test-token"
	defer client.Close()

	err = client.DeleteData("test-id")
	if err != nil {
		t.Errorf("DeleteData failed: %v", err)
	}
}

func TestClientDownloadData(t *testing.T) {
	testData := []byte("test download data")

	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		header, _ := protocol.DeserializeHeader(headerBuf)
		payload := make([]byte, header.Length)
		conn.Read(payload)

		resp := protocol.DownloadResponse{
			Success: true,
			Data:    testData,
			Message: "Download successful",
		}
		respData, _ := protocol.SerializeDownloadResponse(resp)
		message := protocol.SerializeMessage(protocol.MsgTypeDownloadResponse, 1, respData)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	client.username = "testuser"
	client.token = "test-token"
	defer client.Close()

	data, err := client.DownloadData("test-id")
	if err != nil {
		t.Errorf("DownloadData failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Downloaded data mismatch. Got: %s, Expected: %s",
			string(data), string(testData))
	}
}

func TestClientConnectionError(t *testing.T) {
	client := NewClient("invalid-host", 9999)

	// Попытка подключения к несуществующему хосту
	err := client.Connect()
	if err == nil {
		t.Error("Should fail with invalid host")
		client.Close()
	}

	// Все операции должны падать с ошибкой соединения
	err = client.Register("test", "pass")
	if err == nil {
		t.Error("Register should fail without connection")
	}
}

func TestClientJSONSerialization(t *testing.T) {
	// Тестируем обработку невалидного JSON
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		// Отправляем невалидный JSON
		invalidJSON := []byte("{invalid json")
		message := protocol.SerializeMessage(protocol.MsgTypeAuthResponse, 1, invalidJSON)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	defer client.Close()

	err = client.Login("test", "pass")
	if err == nil {
		t.Error("Should fail with invalid JSON response")
	}
}

func TestClientGetData(t *testing.T) {
	testItem := protocol.DataItem{
		ID:   "test-id",
		Type: protocol.DataTypeText,
		Name: "Test Item",
		Data: []byte("test data"),
	}

	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		header, _ := protocol.DeserializeHeader(headerBuf)
		payload := make([]byte, header.Length)
		conn.Read(payload)

		// Проверяем запрос
		var req protocol.DataRequest
		json.Unmarshal(payload, &req)
		if req.ItemID != "test-id" {
			t.Errorf("ItemID mismatch in request: %s", req.ItemID)
		}

		// Отправляем ответ
		resp := protocol.DataResponse{Item: testItem}
		respData, _ := json.Marshal(resp)
		message := protocol.SerializeMessage(protocol.MsgTypeDataResponse, 1, respData)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	client.username = "testuser"
	client.token = "test-token"
	defer client.Close()

	item, err := client.GetData("test-id")
	if err != nil {
		t.Errorf("GetData failed: %v", err)
	}

	if item.ID != "test-id" {
		t.Errorf("Item ID mismatch. Got: %s, Expected: test-id", item.ID)
	}

	if item.Name != "Test Item" {
		t.Errorf("Item name mismatch. Got: %s, Expected: Test Item", item.Name)
	}
}

func TestClientConnectError(t *testing.T) {
	client := NewClient("invalid-hostname-that-does-not-exist", 9999)

	err := client.Connect()
	if err == nil {
		t.Error("Connect should fail with invalid host")
		client.Close()
	}
}

func TestClientSendAndReceiveErrorCases(t *testing.T) {
	// Тест с закрытым соединением
	client := NewClient("localhost", 8080)

	// Пытаемся отправить данные без соединения
	_, err := client.sendAndReceive(protocol.MsgTypeAuthRequest, []byte("test"))
	if err == nil {
		t.Error("Should fail without connection")
	}

	// Тест с разрывом соединения во время чтения
	server := NewMockServer(func(conn net.Conn) {
		conn.Close() // закрываем сразу после подключения
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client2 := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client2.conn = conn

	// Сервер закроет соединение, что вызовет ошибку чтения
	_, err = client2.sendAndReceive(protocol.MsgTypeAuthRequest, []byte("test"))
	if err == nil {
		t.Error("Should fail with closed connection")
	}
}

func TestClientNetworkErrors(t *testing.T) {
	// Тест с таймаутом чтения
	server := NewMockServer(func(conn net.Conn) {
		// Не отправляем ответ, вызывая таймаут
		time.Sleep(100 * time.Millisecond)
		conn.Close()
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Устанавливаем короткий таймаут
	conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	client.conn = conn

	_, err = client.sendAndReceive(protocol.MsgTypeAuthRequest, []byte("test"))
	if err == nil {
		t.Error("Should fail with read timeout")
	}
}

func TestClientPartialRead(t *testing.T) {
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		// Отправляем только часть заголовка
		partialHeader := []byte{0x01, 0x01, 0x00, 0x00}
		conn.Write(partialHeader)
		time.Sleep(10 * time.Millisecond) // даем время на чтение
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn

	_, err = client.sendAndReceive(protocol.MsgTypeAuthRequest, []byte("test"))
	if err == nil {
		t.Error("Should fail with partial read")
	}
}

func TestClientInvalidResponseType(t *testing.T) {
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		// Отправляем ответ с неожиданным типом
		resp := protocol.AuthResponse{Success: true, Token: "test"}
		respData, _ := protocol.SerializeAuthResponse(resp)
		message := protocol.SerializeMessage(0xFF, 1, respData) // неизвестный тип
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	client.username = "testuser"
	client.token = "test-token"

	_, err = client.SyncData(time.Time{})
	if err == nil {
		t.Error("Should fail with unexpected response type")
	}
}

func TestClientEdgeCases(t *testing.T) {
	// Тест с пустыми данными
	server := NewMockServer(func(conn net.Conn) {
		defer conn.Close()

		headerBuf := make([]byte, 10)
		conn.Read(headerBuf)

		// Корректный пустой ответ SyncResponse
		emptyResponse := protocol.SyncResponse{Items: []protocol.DataItem{}}
		respData, err := protocol.SerializeSyncResponse(emptyResponse)
		if err != nil {
			t.Errorf("SerializeSyncResponse failed: %v", err)
			return
		}

		message := protocol.SerializeMessage(protocol.MsgTypeSyncResponse, 1, respData)
		conn.Write(message)
	})

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start mock server: %v", err)
	}
	defer server.Stop()

	client := NewClient("localhost", 0)
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	client.conn = conn
	client.username = "testuser"
	client.token = "test-token"

	items, err := client.SyncData(time.Time{})
	if err != nil {
		t.Errorf("SyncData with empty response failed: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected empty items, got %d", len(items))
	}
}
