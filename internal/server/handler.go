// Package handler предоставляет обработчики клиентских соединений для менеджера паролей.
//
// Обработчики реализуют:
// - Разбор и валидацию входящих сообщений
// - Маршрутизацию запросов к соответствующим методам базы данных
// - Формирование ответов согласно протоколу
// - Управление сессиями и аутентификацией
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"password-manager/internal/common/protocol"
)

// ClientHandler обрабатывает соединение с клиентом.
// Управляет состоянием сессии, аутентификацией и обработкой запросов.
type ClientHandler struct {
	conn       net.Conn
	db         *Database
	username   string
	userID     int
	messageID  uint32
	messageMux sync.Mutex
}

// NewClientHandler создает новый обработчик для клиентского соединения.
//
// Parameters:
//
//	conn - сетевое соединение с клиентом
//	db   - подключение к базе данных
//
// Returns:
//
//	*ClientHandler - новый экземпляр обработчика
//
// Example:
//
//	handler := NewClientHandler(conn, database)
//	go handler.Handle()
func NewClientHandler(conn net.Conn, db *Database) *ClientHandler {
	return &ClientHandler{
		conn: conn,
		db:   db,
	}
}

// Handle обрабатывает входящие сообщения от клиента.
//
// Метод работает в цикле, читая и обрабатывая сообщения до закрытия соединения.
// Автоматически закрывает соединение при завершении работы.
func (h *ClientHandler) Handle() {
	defer h.conn.Close()

	buffer := make([]byte, 50*1024*1024)

	for {
		n, err := h.conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}

		if n < 10 {
			log.Printf("Received message too short: %d bytes", n)
			h.sendError("Message too short")
			continue
		}

		header, payload, err := protocol.DeserializeMessage(buffer[:n])
		if err != nil {
			log.Printf("Error deserializing message: %v", err)
			h.sendError("Invalid message format")
			continue
		}

		log.Printf("Received message type: %d, length: %d", header.Type, header.Length)

		h.handleMessage(header.Type, payload)
	}
}

// handleMessage маршрутизирует входящее сообщение к соответствующему обработчику.
//
// Parameters:
//
//	msgType - тип сообщения из протокола
//	data    - данные сообщения
func (h *ClientHandler) handleMessage(msgType uint8, data []byte) {
	switch msgType {
	case protocol.MsgTypeAuthRequest:
		h.handleAuthRequest(data)
	case protocol.MsgTypeRegisterRequest:
		h.handleRegisterRequest(data)
	case protocol.MsgTypeSyncRequest:
		h.handleSyncRequest(data)
	case protocol.MsgTypeDataRequest:
		h.handleDataRequest(data)
	case protocol.MsgTypeSaveDataRequest:
		h.handleSaveDataRequest(data)
	case protocol.MsgTypeDeleteDataRequest:
		h.handleDeleteDataRequest(data)
	case protocol.MsgTypeUpdateDataRequest:
		h.handleUpdateDataRequest(data)
	case protocol.MsgTypeDownloadRequest:
		h.handleDownloadRequest(data)
	default:
		h.sendError("Unknown message type")
	}
}

// handleAuthRequest обрабатывает запрос аутентификации.
//
// Parameters:
//
//	data - данные запроса в формате AuthRequest
func (h *ClientHandler) handleAuthRequest(data []byte) {
	req, err := protocol.DeserializeAuthRequest(data)
	if err != nil {
		log.Printf("Error deserializing auth request: %v", err)
		h.sendError("Invalid auth request format")
		return
	}

	log.Printf("Auth request for user: %s", req.Username)

	authenticated, err := h.db.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		log.Printf("Authentication error: %v", err)
		h.sendError("Authentication error")
		return
	}

	if authenticated {
		h.username = req.Username
		userID, err := h.db.GetUserID(req.Username)
		if err != nil {
			log.Printf("Error getting user ID: %v", err)
			h.sendError("User not found")
			return
		}
		h.userID = userID

		resp := protocol.AuthResponse{
			Success: true,
			Token:   "dummy-token",
		}
		h.sendResponse(protocol.MsgTypeAuthResponse, resp)
		log.Printf("User %s authenticated successfully", req.Username)
	} else {
		h.sendError("Authentication failed: invalid credentials")
		log.Printf("Authentication failed for user: %s", req.Username)
	}
}

// handleRegisterRequest обрабатывает запрос регистрации нового пользователя.
//
// Parameters:
//
//	data - данные запроса в формате RegisterRequest
//
// Process:
//   - Десериализует запрос регистрации
//   - Создает нового пользователя в базе данных
//   - Отправляет ответ с результатом регистрации
//   - Логирует успешную регистрацию или ошибки
func (h *ClientHandler) handleRegisterRequest(data []byte) {
	req, err := protocol.DeserializeRegisterRequest(data)
	if err != nil {
		log.Printf("Error deserializing register request: %v", err)
		h.sendError("Invalid register request format")
		return
	}

	log.Printf("Register request for user: %s", req.Username)

	err = h.db.CreateUser(req.Username, req.Password)
	if err != nil {
		log.Printf("Registration error: %v", err)
		h.sendError(fmt.Sprintf("Registration failed: %v", err))
		return
	}

	resp := protocol.RegisterResponse{
		Success: true,
		Message: "User registered successfully",
	}
	h.sendResponse(protocol.MsgTypeRegisterResponse, resp)
	log.Printf("User %s registered successfully", req.Username)
}

// handleSyncRequest обрабатывает запрос синхронизации данных.
//
// Parameters:
//
//	data - данные запроса в формате SyncRequest
func (h *ClientHandler) handleSyncRequest(data []byte) {
	if h.userID == 0 {
		h.sendError("Not authenticated")
		return
	}

	req, err := protocol.DeserializeSyncRequest(data)
	if err != nil {
		log.Printf("Error deserializing sync request: %v", err)
		h.sendError("Invalid sync request format")
		return
	}

	log.Printf("Sync request from user %s, last sync: %v", h.username, req.LastSync)

	items, err := h.db.GetData(h.userID, req.LastSync)
	if err != nil {
		log.Printf("Error getting data: %v", err)
		h.sendError("Failed to get data")
		return
	}

	resp := protocol.SyncResponse{Items: items}
	h.sendResponse(protocol.MsgTypeSyncResponse, resp)
	log.Printf("Sent %d items to user %s", len(items), h.username)
}

// handleDataRequest обрабатывает запрос конкретного элемента данных.
//
// Parameters:
//
//	data - данные запроса в формате DataRequest
func (h *ClientHandler) handleDataRequest(data []byte) {
	if h.userID == 0 {
		h.sendError("Not authenticated")
		return
	}

	var req protocol.DataRequest
	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("Error deserializing data request: %v", err)
		h.sendError("Invalid data request format")
		return
	}

	log.Printf("Data request from user %s for item: %s", h.username, req.ItemID)

	item, err := h.db.GetDataByID(h.userID, req.ItemID)
	if err != nil {
		log.Printf("Error getting data by ID: %v", err)
		h.sendError("Data not found")
		return
	}

	resp := protocol.DataResponse{Item: item}
	h.sendResponse(protocol.MsgTypeDataResponse, resp)
	log.Printf("Sent data item %s to user %s", req.ItemID, h.username)
}

// handleSaveDataRequest обрабатывает запрос сохранения данных.
//
// Parameters:
//
//	data - данные запроса в формате SaveDataRequest
func (h *ClientHandler) handleSaveDataRequest(data []byte) {
	if h.userID == 0 {
		h.sendError("Not authenticated")
		return
	}

	req, err := protocol.DeserializeSaveDataRequest(data)
	if err != nil {
		log.Printf("Error deserializing save data request: %v", err)
		h.sendError("Invalid save data request format")
		return
	}

	log.Printf("Save data request from user %s for item: %s", h.username, req.Item.Name)

	err = h.db.StoreData(h.userID, req.Item)
	if err != nil {
		log.Printf("Error saving data: %v", err)
		h.sendError(fmt.Sprintf("Failed to store data: %v", err))
		return
	}

	resp := protocol.SaveDataResponse{
		Success: true,
		Message: "Data saved successfully",
		ItemID:  "",
	}
	h.sendResponse(protocol.MsgTypeSaveDataResponse, resp)
	log.Printf("Saved data for user %s: %s", h.username, req.Item.Name)
}

// sendResponse отправляет ответ клиенту.
//
// Parameters:
//
//	msgType - тип ответного сообщения
//	data    - данные для отправки (интерфейс, сериализуемый в JSON)
func (h *ClientHandler) sendResponse(msgType uint8, data interface{}) {
	h.messageMux.Lock()
	defer h.messageMux.Unlock()

	var serialized []byte
	var err error

	switch v := data.(type) {
	case protocol.AuthResponse:
		serialized, err = protocol.SerializeAuthResponse(v)
	case protocol.RegisterResponse:
		serialized, err = protocol.SerializeRegisterResponse(v)
	case protocol.SyncResponse:
		serialized, err = protocol.SerializeSyncResponse(v)
	case protocol.DataResponse:
		serialized, err = json.Marshal(v)
	case protocol.SaveDataResponse:
		serialized, err = protocol.SerializeSaveDataResponse(v)
	case protocol.DeleteDataResponse:
		serialized, err = protocol.SerializeDeleteDataResponse(v)
	case protocol.UpdateDataResponse:
		serialized, err = protocol.SerializeUpdateDataResponse(v)
	case protocol.DownloadResponse:
		serialized, err = protocol.SerializeDownloadResponse(v)
	default:
		h.sendError("Unknown response type")
		return
	}

	if err != nil {
		log.Printf("Error serializing response: %v", err)
		h.sendError("Failed to serialize response")
		return
	}

	message := protocol.SerializeMessage(msgType, h.messageID, serialized)
	h.messageID++

	_, err = h.conn.Write(message)
	if err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

// sendError отправляет сообщение об ошибке клиенту.
//
// Parameters:
//
//	message - текст ошибки
func (h *ClientHandler) sendError(message string) {
	h.messageMux.Lock()
	defer h.messageMux.Unlock()

	errorResp := protocol.ErrorResponse{
		Code:    500,
		Message: message,
	}

	serialized, err := protocol.SerializeErrorResponse(errorResp)
	if err != nil {
		log.Printf("Failed to serialize error: %v", err)
		return
	}

	messageData := protocol.SerializeMessage(protocol.MsgTypeError, h.messageID, serialized)
	h.messageID++

	_, err = h.conn.Write(messageData)
	if err != nil {
		log.Printf("Error sending error: %v", err)
	}
}

// handleUpdateDataRequest обрабатывает запрос обновления элемента данных.
//
// Parameters:
//
//	data - данные запроса в формате UpdateDataRequest
func (h *ClientHandler) handleDeleteDataRequest(data []byte) {
	if h.userID == 0 {
		h.sendError("Not authenticated")
		return
	}

	req, err := protocol.DeserializeDeleteDataRequest(data)
	if err != nil {
		log.Printf("Error deserializing delete data request: %v", err)
		h.sendError("Invalid delete data request format")
		return
	}

	log.Printf("Delete data request from user %s for item: %s", h.username, req.ItemID)

	err = h.db.DeleteData(h.userID, req.ItemID)
	if err != nil {
		log.Printf("Error deleting data: %v", err)
		h.sendError(fmt.Sprintf("Failed to delete data: %v", err))
		return
	}

	resp := protocol.DeleteDataResponse{
		Success: true,
		Message: "Data deleted successfully",
	}
	h.sendResponse(protocol.MsgTypeDeleteDataResponse, resp)
	log.Printf("Deleted data for user %s: %s", h.username, req.ItemID)
}

// handleUpdateDataRequest обрабатывает запрос обновления элемента данных.
//
// Parameters:
//
//	data - данные запроса в формате UpdateDataRequest
func (h *ClientHandler) handleUpdateDataRequest(data []byte) {
	if h.userID == 0 {
		h.sendError("Not authenticated")
		return
	}

	req, err := protocol.DeserializeUpdateDataRequest(data)
	if err != nil {
		log.Printf("Error deserializing update data request: %v", err)
		h.sendError("Invalid update data request format")
		return
	}

	log.Printf("Update data request from user %s for item: %s", h.username, req.ItemID)

	err = h.db.UpdateData(h.userID, req.ItemID, req.Item)
	if err != nil {
		log.Printf("Error updating data: %v", err)
		h.sendError(fmt.Sprintf("Failed to update data: %v", err))
		return
	}

	resp := protocol.UpdateDataResponse{
		Success: true,
		Message: "Data updated successfully",
	}
	h.sendResponse(protocol.MsgTypeUpdateDataResponse, resp)
	log.Printf("Updated data for user %s: %s", h.username, req.ItemID)
}

// handleDownloadRequest обрабатывает запрос загрузки данных элемента.
//
// Parameters:
//
//	data - данные запроса в формате DownloadRequest
func (h *ClientHandler) handleDownloadRequest(data []byte) {
	if h.userID == 0 {
		h.sendError("Not authenticated")
		return
	}

	req, err := protocol.DeserializeDownloadRequest(data)
	if err != nil {
		log.Printf("Error deserializing download request: %v", err)
		h.sendError("Invalid download request format")
		return
	}

	log.Printf("Download request from user %s for item: %s", h.username, req.ItemID)

	item, err := h.db.GetDataByID(h.userID, req.ItemID)
	if err != nil {
		log.Printf("Error getting data by ID: %v", err)
		h.sendError("Data not found")
		return
	}

	resp := protocol.DownloadResponse{
		Success: true,
		Data:    item.Data,
		Message: "Download successful",
	}
	h.sendResponse(protocol.MsgTypeDownloadResponse, resp)
	log.Printf("Sent download data for user %s: %s (%d bytes)", h.username, req.ItemID, len(item.Data))
}
