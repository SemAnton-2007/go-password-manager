// Package client предоставляет клиентскую библиотеку для взаимодействия с сервером менеджера паролей.
//
// Клиент реализует:
// - Установку и поддержание соединения с сервером
// - Сериализацию/десериализацию сообщений протокола
// - Управление аутентификацией и сессиями
// - Операции с данными (создание, чтение, обновление, удаление)
//
// Пример использования:
//
//	cl := client.NewClient("localhost", 8080)
//	err := cl.Connect()
//	err := cl.Login("username", "password")
package client

import (
	"fmt"
	"io"
	"net"
	"time"

	"password-manager/internal/common/protocol"
)

type Client struct {
	conn     net.Conn
	host     string
	port     int
	token    string
	username string
}

// NewClient создает новый клиент для подключения к серверу.
//
// Parameters:
//
//	host - хост сервера
//	port - порт сервера
//
// Returns:
//
//	*Client - новый экземпляр клиента
func NewClient(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

// Connect устанавливает TCP соединение с сервером.
//
// Returns:
//
//	error - ошибка если соединение не удалось установить
func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	c.conn = conn
	return nil
}

// Close закрывает соединение с сервером.
//
// Returns:
//
//	error - ошибка закрытия соединения
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) sendAndReceive(msgType uint8, data []byte) ([]byte, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	message := protocol.SerializeMessage(msgType, 1, data)
	_, err := c.conn.Write(message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	headerBuf := make([]byte, 10)
	_, err = io.ReadFull(c.conn, headerBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	header, err := protocol.DeserializeHeader(headerBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	payload := make([]byte, header.Length)
	_, err = io.ReadFull(c.conn, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	if header.Type == protocol.MsgTypeError {
		errorResp, err := protocol.DeserializeErrorResponse(payload)
		if err != nil {
			return nil, fmt.Errorf("error response: failed to parse: %w", err)
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	return payload, nil
}

// Register регистрирует нового пользователя на сервере.
//
// Parameters:
//
//	username - имя пользователя
//	password - пароль
//
// Returns:
//
//	error - ошибка если регистрация не удалась
func (c *Client) Register(username, password string) error {
	req := protocol.RegisterRequest{
		Username: username,
		Password: password,
	}

	data, err := protocol.SerializeRegisterRequest(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %w", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeRegisterRequest, data)
	if err != nil {
		return err
	}

	resp, err := protocol.DeserializeRegisterResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	return nil
}

// Login выполняет аутентификацию пользователя.
//
// Parameters:
//
//	username - имя пользователя
//	password - пароль
//
// Returns:
//
//	error - ошибка если аутентификация не удалась
func (c *Client) Login(username, password string) error {
	req := protocol.AuthRequest{
		Username: username,
		Password: password,
	}

	data, err := protocol.SerializeAuthRequest(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %w", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeAuthRequest, data)
	if err != nil {
		return err
	}

	resp, err := protocol.DeserializeAuthResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("authentication failed")
	}

	c.token = resp.Token
	c.username = username
	return nil
}

// SyncData синхронизирует данные с сервером.
//
// Parameters:
//
//	lastSync - время последней успешной синхронизации
//
// Returns:
//
//	[]DataItem - список измененных элементов
//	error - ошибка синхронизации
func (c *Client) SyncData(lastSync time.Time) ([]protocol.DataItem, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	req := protocol.SyncRequest{
		LastSync: lastSync,
	}

	data, err := protocol.SerializeSyncRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeSyncRequest, data)
	if err != nil {
		return nil, err
	}

	resp, err := protocol.DeserializeSyncResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp.Items, nil
}

// SaveData сохраняет новый элемент данных на сервере.
//
// Parameters:
//
//	item - элемент данных для сохранения
//
// Returns:
//
//	error - ошибка сохранения
func (c *Client) SaveData(item protocol.NewDataItem) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	req := protocol.SaveDataRequest{
		Item: item,
	}

	data, err := protocol.SerializeSaveDataRequest(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %w", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeSaveDataRequest, data)
	if err != nil {
		return err
	}

	resp, err := protocol.DeserializeSaveDataResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to save data: %s", resp.Message)
	}

	return nil
}

// IsAuthenticated проверяет статус аутентификации клиента.
//
// Returns:
//
//	bool - true если клиент аутентифицирован
func (c *Client) IsAuthenticated() bool {
	return c.token != "" && c.username != ""
}

// GetUsername возвращает имя текущего аутентифицированного пользователя.
//
// Returns:
//
//	string - имя пользователя или пустая строка если не аутентифицирован
func (c *Client) GetUsername() string {
	return c.username
}

// DeleteData удаляет элемент данных с сервера.
//
// Parameters:
//
//	itemID - ID элемента для удаления
//
// Returns:
//
//	error - ошибка удаления
func (c *Client) DeleteData(itemID string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	req := protocol.DeleteDataRequest{
		ItemID: itemID,
	}

	data, err := protocol.SerializeDeleteDataRequest(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %w", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeDeleteDataRequest, data)
	if err != nil {
		return err
	}

	resp, err := protocol.DeserializeDeleteDataResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to delete data: %s", resp.Message)
	}

	return nil
}

// UpdateData обновляет существующий элемент данных на сервере.
//
// Parameters:
//
//	itemID - ID элемента для обновления
//	item   - новые данные элемента
//
// Returns:
//
//	error - ошибка обновления
func (c *Client) UpdateData(itemID string, item protocol.NewDataItem) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	req := protocol.UpdateDataRequest{
		ItemID: itemID,
		Item:   item,
	}

	data, err := protocol.SerializeUpdateDataRequest(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %w", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeUpdateDataRequest, data)
	if err != nil {
		return err
	}

	resp, err := protocol.DeserializeUpdateDataResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to update data: %s", resp.Message)
	}

	return nil
}

// DownloadData загружает данные элемента
//
// Parameters:
//
//	itemID - ID элемента для загрузки
//
// Returns:
//
//	[]byte - загруженные данные
//	error  - ошибка загрузки
func (c *Client) DownloadData(itemID string) ([]byte, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	req := protocol.DownloadRequest{
		ItemID: itemID,
	}

	data, err := protocol.SerializeDownloadRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeDownloadRequest, data)
	if err != nil {
		return nil, err
	}

	resp, err := protocol.DeserializeDownloadResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to download data: %s", resp.Message)
	}

	return resp.Data, nil
}
