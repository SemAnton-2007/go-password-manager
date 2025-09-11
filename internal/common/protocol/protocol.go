// Package protocol определяет сетевой протокол для менеджера паролей.
//
// Протокол включает:
// - Форматы сообщений для всех операций
// - Типы данных и их сериализацию
// - Коды ошибок и статусные сообщения
// - Поддержку метаданных для всех элементов
//
// Сообщения используют бинарный формат с заголовком фиксированной длины
// и телом переменной длины в формате JSON.
package protocol

import (
	"encoding/binary"
	"encoding/json"
	"time"
)

// SerializeMessage создает бинарное сообщение из заголовка и данных.
//
// Parameters:
//
//	msgType   - тип сообщения
//	messageID - уникальный ID сообщения
//	data      - данные сообщения
//
// Returns:
//
//	[]byte - сериализованное сообщение
//
// Format:
//
//	[0:1]  - тип сообщения
//	[1:2]  - версия протокола
//	[2:6]  - ID сообщения (uint32 big endian)
//	[6:10] - длина данных (uint32 big endian)
//	[10:]  - данные сообщения
func SerializeMessage(msgType uint8, messageID uint32, data []byte) []byte {
	header := MessageHeader{
		Type:      msgType,
		Version:   1,
		MessageID: messageID,
		Length:    uint32(len(data)),
	}

	buf := make([]byte, 10)
	buf[0] = header.Type
	buf[1] = header.Version
	binary.BigEndian.PutUint32(buf[2:6], header.MessageID)
	binary.BigEndian.PutUint32(buf[6:10], header.Length)

	return append(buf, data...)
}

// DeserializeMessage разбирает бинарное сообщение на заголовок и данные.
//
// Parameters:
//
//	data - бинарное сообщение
//
// Returns:
//
//	MessageHeader - разобранный заголовок
//	[]byte        - данные сообщения
//	error         - ошибка если сообщение невалидно
func DeserializeMessage(data []byte) (MessageHeader, []byte, error) {
	if len(data) < 10 {
		return MessageHeader{}, nil, ErrInvalidMessage
	}

	header := MessageHeader{
		Type:      data[0],
		Version:   data[1],
		MessageID: binary.BigEndian.Uint32(data[2:6]),
		Length:    binary.BigEndian.Uint32(data[6:10]),
	}

	// Если данных меньше чем заголовок + payload, возвращаем только заголовок
	if len(data) < 10+int(header.Length) {
		return header, nil, nil
	}

	return header, data[10 : 10+header.Length], nil
}

// SerializeAuthRequest сериализует запрос аутентификации в JSON.
//
// Parameters:
//
//	req - структура запроса аутентификации
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeAuthRequest(req AuthRequest) ([]byte, error) {
	return json.Marshal(req)
}

// DeserializeAuthRequest десериализует запрос аутентификации из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	AuthRequest - разобранная структура
//	error       - ошибка десериализации
func DeserializeAuthRequest(data []byte) (AuthRequest, error) {
	var req AuthRequest
	err := json.Unmarshal(data, &req)
	return req, err
}

// SerializeAuthResponse сериализует ответ аутентификации в JSON.
//
// Parameters:
//
//	resp - структура ответа аутентификации
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeAuthResponse(resp AuthResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// DeserializeAuthResponse десериализует ответ аутентификации из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	AuthResponse - разобранная структура
//	error        - ошибка десериализации
func DeserializeAuthResponse(data []byte) (AuthResponse, error) {
	var resp AuthResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

// SerializeRegisterRequest сериализует запрос регистрации в JSON.
//
// Parameters:
//
//	req - структура запроса регистрации
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeRegisterRequest(req RegisterRequest) ([]byte, error) {
	return json.Marshal(req)
}

// DeserializeRegisterRequest десериализует запрос регистрации из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	RegisterRequest - разобранная структура
//	error           - ошибка десериализации
func DeserializeRegisterRequest(data []byte) (RegisterRequest, error) {
	var req RegisterRequest
	err := json.Unmarshal(data, &req)
	return req, err
}

// SerializeRegisterResponse сериализует ответ регистрации в JSON.
//
// Parameters:
//
//	resp - структура ответа регистрации
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeRegisterResponse(resp RegisterResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// DeserializeRegisterResponse десериализует ответ регистрации из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	RegisterResponse - разобранная структура
//	error            - ошибка десериализации
func DeserializeRegisterResponse(data []byte) (RegisterResponse, error) {
	var resp RegisterResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

// SerializeSyncRequest сериализует запрос синхронизации в JSON.
//
// Parameters:
//
//	req - структура запроса синхронизации
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeSyncRequest(req SyncRequest) ([]byte, error) {
	return json.Marshal(struct {
		LastSync string `json:"last_sync"`
	}{
		LastSync: req.LastSync.Format(time.RFC3339Nano),
	})
}

// DeserializeSyncRequest десериализует запрос синхронизации из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	SyncRequest - разобранная структура
//	error       - ошибка десериализации
func DeserializeSyncRequest(data []byte) (SyncRequest, error) {
	var temp struct {
		LastSync string `json:"last_sync"`
	}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return SyncRequest{}, err
	}

	var lastSync time.Time
	if temp.LastSync != "" {
		lastSync, err = time.Parse(time.RFC3339Nano, temp.LastSync)
		if err != nil {
			return SyncRequest{}, err
		}
	}

	return SyncRequest{LastSync: lastSync}, nil
}

// SerializeSyncResponse сериализует ответ синхронизации в JSON.
//
// Parameters:
//
//	resp - структура ответа синхронизации
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeSyncResponse(resp SyncResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// DeserializeSyncResponse десериализует ответ синхронизации из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	SyncResponse - разобранная структура
//	error        - ошибка десериализации
func DeserializeSyncResponse(data []byte) (SyncResponse, error) {
	var resp SyncResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

// SerializeDataItem сериализует элемент данных в JSON.
//
// Parameters:
//
//	item - элемент данных для сериализации
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeDataItem(item DataItem) ([]byte, error) {
	type dataItem struct {
		ID        string            `json:"id"`
		Type      uint8             `json:"type"`
		Name      string            `json:"name"`
		Data      []byte            `json:"data"`
		Metadata  map[string]string `json:"metadata"`
		CreatedAt string            `json:"created_at"`
		UpdatedAt string            `json:"updated_at"`
	}

	temp := dataItem{
		ID:        item.ID,
		Type:      item.Type,
		Name:      item.Name,
		Data:      item.Data,
		Metadata:  item.Metadata,
		CreatedAt: item.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339Nano),
	}

	return json.Marshal(temp)
}

// DeserializeDataItem десериализует элемент данных из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	DataItem - разобранный элемент данных
//	error    - ошибка десериализации
func DeserializeDataItem(data []byte) (DataItem, error) {
	type dataItem struct {
		ID        string            `json:"id"`
		Type      uint8             `json:"type"`
		Name      string            `json:"name"`
		Data      []byte            `json:"data"`
		Metadata  map[string]string `json:"metadata"`
		CreatedAt string            `json:"created_at"`
		UpdatedAt string            `json:"updated_at"`
	}

	var temp dataItem
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return DataItem{}, err
	}

	createdAt, err := time.Parse(time.RFC3339Nano, temp.CreatedAt)
	if err != nil {
		return DataItem{}, err
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, temp.UpdatedAt)
	if err != nil {
		return DataItem{}, err
	}

	return DataItem{
		ID:        temp.ID,
		Type:      temp.Type,
		Name:      temp.Name,
		Data:      temp.Data,
		Metadata:  temp.Metadata,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// SerializeSaveDataRequest сериализует запрос сохранения данных в JSON.
//
// Parameters:
//
//	req - структура запроса сохранения данных
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeSaveDataRequest(req SaveDataRequest) ([]byte, error) {
	type tempDataItem struct {
		Type     uint8             `json:"type"`
		Name     string            `json:"name"`
		Data     []byte            `json:"data"`
		Metadata map[string]string `json:"metadata"`
	}

	type tempRequest struct {
		Item tempDataItem `json:"item"`
	}

	temp := tempDataItem{
		Type:     req.Item.Type,
		Name:     req.Item.Name,
		Data:     req.Item.Data,
		Metadata: req.Item.Metadata,
	}

	return json.Marshal(tempRequest{Item: temp})
}

// DeserializeSaveDataRequest десериализует запрос сохранения данных из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	SaveDataRequest - разобранная структура
//	error           - ошибка десериализации
func DeserializeSaveDataRequest(data []byte) (SaveDataRequest, error) {
	type tempDataItem struct {
		Type     uint8             `json:"type"`
		Name     string            `json:"name"`
		Data     []byte            `json:"data"`
		Metadata map[string]string `json:"metadata"`
	}

	type tempRequest struct {
		Item tempDataItem `json:"item"`
	}

	var temp tempRequest
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return SaveDataRequest{}, err
	}

	return SaveDataRequest{
		Item: NewDataItem{
			Type:     temp.Item.Type,
			Name:     temp.Item.Name,
			Data:     temp.Item.Data,
			Metadata: temp.Item.Metadata,
		},
	}, nil
}

// SerializeSaveDataResponse сериализует ответ сохранения данных в JSON.
//
// Parameters:
//
//	resp - структура ответа сохранения
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeSaveDataResponse(resp SaveDataResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// DeserializeSaveDataResponse десериализует ответ сохранения данных из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	SaveDataResponse - разобранная структура
//	error            - ошибка десериализации
func DeserializeSaveDataResponse(data []byte) (SaveDataResponse, error) {
	var resp SaveDataResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

// SerializeErrorResponse сериализует ответ с ошибкой в JSON.
//
// Parameters:
//
//	resp - структура ответа с ошибкой
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeErrorResponse(resp ErrorResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// DeserializeErrorResponse десериализует ответ с ошибкой из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	ErrorResponse - разобранная структура
//	error         - ошибка десериализации
func DeserializeErrorResponse(data []byte) (ErrorResponse, error) {
	var resp ErrorResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

// SerializeDeleteDataRequest сериализует запрос удаления данных в JSON.
//
// Parameters:
//
//	req - структура запроса удаления данных
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeDeleteDataRequest(req DeleteDataRequest) ([]byte, error) {
	return json.Marshal(req)
}

// DeserializeDeleteDataRequest десериализует запрос удаления данных из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	DeleteDataRequest - разобранная структура
//	error             - ошибка десериализации
func DeserializeDeleteDataRequest(data []byte) (DeleteDataRequest, error) {
	var req DeleteDataRequest
	err := json.Unmarshal(data, &req)
	return req, err
}

// SerializeDeleteDataResponse сериализует ответ удаления данных в JSON.
//
// Parameters:
//
//	resp - структура ответа удаления данных
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeDeleteDataResponse(resp DeleteDataResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// DeserializeDeleteDataResponse десериализует ответ удаления данных из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	DeleteDataResponse - разобранная структура
//	error              - ошибка десериализации
func DeserializeDeleteDataResponse(data []byte) (DeleteDataResponse, error) {
	var resp DeleteDataResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

// SerializeUpdateDataRequest сериализует запрос обновления данных в JSON.
//
// Parameters:
//
//	req - структура запроса обновления данных
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeUpdateDataRequest(req UpdateDataRequest) ([]byte, error) {
	type tempDataItem struct {
		Type     uint8             `json:"type"`
		Name     string            `json:"name"`
		Data     []byte            `json:"data"`
		Metadata map[string]string `json:"metadata"`
	}

	type tempRequest struct {
		ItemID string       `json:"item_id"`
		Item   tempDataItem `json:"item"`
	}

	temp := tempRequest{
		ItemID: req.ItemID,
		Item: tempDataItem{
			Type:     req.Item.Type,
			Name:     req.Item.Name,
			Data:     req.Item.Data,
			Metadata: req.Item.Metadata,
		},
	}

	return json.Marshal(temp)
}

// DeserializeUpdateDataRequest десериализует запрос обновления данных из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	UpdateDataRequest - разобранная структура
//	error             - ошибка десериализации
func DeserializeUpdateDataRequest(data []byte) (UpdateDataRequest, error) {
	type tempDataItem struct {
		Type     uint8             `json:"type"`
		Name     string            `json:"name"`
		Data     []byte            `json:"data"`
		Metadata map[string]string `json:"metadata"`
	}

	type tempRequest struct {
		ItemID string       `json:"item_id"`
		Item   tempDataItem `json:"item"`
	}

	var temp tempRequest
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return UpdateDataRequest{}, err
	}

	return UpdateDataRequest{
		ItemID: temp.ItemID,
		Item: NewDataItem{
			Type:     temp.Item.Type,
			Name:     temp.Item.Name,
			Data:     temp.Item.Data,
			Metadata: temp.Item.Metadata,
		},
	}, nil
}

// SerializeUpdateDataResponse сериализует ответ обновления данных в JSON.
//
// Parameters:
//
//	resp - структура ответа обновления данных
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeUpdateDataResponse(resp UpdateDataResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// DeserializeUpdateDataResponse десериализует ответ обновления данных из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	UpdateDataResponse - разобранная структура
//	error              - ошибка десериализации
func DeserializeUpdateDataResponse(data []byte) (UpdateDataResponse, error) {
	var resp UpdateDataResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

// SerializeDownloadRequest сериализует запрос загрузки данных в JSON.
//
// Parameters:
//
//	req - структура запроса загрузки данных
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeDownloadRequest(req DownloadRequest) ([]byte, error) {
	return json.Marshal(req)
}

// DeserializeDownloadRequest десериализует запрос загрузки данных из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	DownloadRequest - разобранная структура
//	error           - ошибка десериализации
func DeserializeDownloadRequest(data []byte) (DownloadRequest, error) {
	var req DownloadRequest
	err := json.Unmarshal(data, &req)
	return req, err
}

// SerializeDownloadResponse сериализует ответ загрузки данных в JSON.
//
// Parameters:
//
//	resp - структура ответа загрузки данных
//
// Returns:
//
//	[]byte - сериализованные данные
//	error  - ошибка сериализации
func SerializeDownloadResponse(resp DownloadResponse) ([]byte, error) {
	return json.Marshal(resp)
}

// DeserializeDownloadResponse десериализует ответ загрузки данных из JSON.
//
// Parameters:
//
//	data - сериализованные данные
//
// Returns:
//
//	DownloadResponse - разобранная структура
//	error            - ошибка десериализации
func DeserializeDownloadResponse(data []byte) (DownloadResponse, error) {
	var req DownloadResponse
	err := json.Unmarshal(data, &req)
	return req, err
}

// DeserializeHeader разбирает заголовок сообщения из бинарных данных.
//
// Parameters:
//
//	data - бинарные данные заголовка (минимум 10 байт)
//
// Returns:
//
//	MessageHeader - разобранный заголовок
//	error         - ошибка если данные невалидны
func DeserializeHeader(data []byte) (MessageHeader, error) {
	if len(data) < 10 {
		return MessageHeader{}, ErrInvalidMessage
	}

	return MessageHeader{
		Type:      data[0],
		Version:   data[1],
		MessageID: binary.BigEndian.Uint32(data[2:6]),
		Length:    binary.BigEndian.Uint32(data[6:10]),
	}, nil
}
