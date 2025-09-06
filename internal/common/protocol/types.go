// Package types определяет структуры данных и константы сетевого протокола.
//
// Включает:
// - Типы сообщений и их коды
// - Структуры запросов и ответов
// - Типы хранимых данных
// - Константы метаданных
// - Определения ошибок
//
// Протокол использует бинарный формат с заголовком фиксированной длины
// и JSON-сериализацией для тела сообщений.
package protocol

import (
	"errors"
	"time"
)

// Типы сообщений протокола
const (
	// MsgTypeAuthRequest - запрос аутентификации пользователя
	MsgTypeAuthRequest = 0x01
	// MsgTypeAuthResponse - ответ на запрос аутентификации
	MsgTypeAuthResponse = 0x02
	// MsgTypeRegisterRequest - запрос регистрации нового пользователя
	MsgTypeRegisterRequest = 0x03
	// MsgTypeRegisterResponse - ответ на запрос регистрации
	MsgTypeRegisterResponse = 0x04
	// MsgTypeSyncRequest - запрос синхронизации данных пользователя
	MsgTypeSyncRequest = 0x05
	// MsgTypeSyncResponse - ответ с синхронизированными данными
	MsgTypeSyncResponse = 0x06
	// MsgTypeDataRequest - запрос конкретного элемента данных по ID
	MsgTypeDataRequest = 0x07
	// MsgTypeDataResponse - ответ с запрошенным элементом данных
	MsgTypeDataResponse = 0x08
	// MsgTypeSaveDataRequest - запрос сохранения нового элемента данных
	MsgTypeSaveDataRequest = 0x09
	// MsgTypeSaveDataResponse - ответ на запрос сохранения данных
	MsgTypeSaveDataResponse = 0x0A
	// MsgTypeError - сообщение об ошибке
	MsgTypeError = 0xFF
	// MsgTypeDeleteDataRequest - запрос удаления элемента данных
	MsgTypeDeleteDataRequest = 0x0B
	// MsgTypeDeleteDataResponse - ответ на запрос удаления данных
	MsgTypeDeleteDataResponse = 0x0C
	// MsgTypeUpdateDataRequest - запрос обновления существующего элемента данных
	MsgTypeUpdateDataRequest = 0x0D
	// MsgTypeUpdateDataResponse - ответ на запрос обновления данных
	MsgTypeUpdateDataResponse = 0x0E
	// MsgTypeDownloadRequest - запрос загрузки данных
	MsgTypeDownloadRequest = 0x0F
	// MsgTypeDownloadResponse - ответ с загруженными данными
	MsgTypeDownloadResponse = 0x10
)

// Типы данных, поддерживаемые системой
const (
	// DataTypeLoginPassword - учетные данные (логин/пароль)
	DataTypeLoginPassword = 0x01
	// DataTypeText - произвольные текстовые данные
	DataTypeText = 0x02
	// DataTypeBinary - бинарные данные
	DataTypeBinary = 0x03
	// DataTypeBankCard - данные банковских карт
	DataTypeBankCard = 0x04
)

// Ключи метаданных для бинарных данных
const (
	// MetaOriginalFileName - оригинальное имя файла для бинарных данных
	MetaOriginalFileName = "original_file_name"
	// MetaFileSize - размер файла в байтах
	MetaFileSize = "file_size"
	// MetaFileExtension - расширение файла
	MetaFileExtension = "file_extension"
)

var (
	// ErrInvalidMessage возвращается при получении невалидного сообщения.
	// Обычно указывает на повреждение данных или несовместимость версий.
	ErrInvalidMessage = errors.New("invalid message")
	// ErrAuthFailed возвращается при неудачной аутентификации.
	// Может быть вызвано неверными credentials или блокировкой аккаунта.
	ErrAuthFailed = errors.New("authentication failed")
)

// MessageHeader представляет заголовок сетевого сообщения.
// Содержит метаинформацию о сообщении: тип, версию, ID и длину данных.
type MessageHeader struct {
	Type      uint8
	Version   uint8
	MessageID uint32
	Length    uint32
}

// AuthRequest содержит credentials для аутентификации пользователя.
// Используется при входе в систему.
type AuthRequest struct {
	Username string
	Password string
}

// AuthResponse содержит результат попытки аутентификации.
// Включает статус успеха и токен сессии (если успешно).
type AuthResponse struct {
	Success bool
	Token   string
}

// RegisterRequest содержит данные для регистрации нового пользователя.
// Используется при создании нового аккаунта в системе.
type RegisterRequest struct {
	Username string
	Password string
}

// RegisterResponse содержит результат попытки регистрации пользователя.
// Включает статус успеха и сообщение для пользователя.
type RegisterResponse struct {
	Success bool
	Message string
}

// DataItem представляет элемент данных, хранимый в системе.
// Может содержать различные типы данных с метаинформацией.
type DataItem struct {
	ID        string
	Type      uint8
	Name      string
	Data      []byte
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewDataItem представляет новый элемент данных для создания.
// Используется при добавлении новых записей.
type NewDataItem struct {
	Type     uint8
	Name     string
	Data     []byte
	Metadata map[string]string
}

// SyncRequest содержит запрос на синхронизацию данных.
// LastSync указывает время последней успешной синхронизации.
type SyncRequest struct {
	LastSync time.Time
}

// SyncResponse содержит результаты синхронизации данных.
// Items содержит все элементы, измененные после LastSync.
type SyncResponse struct {
	Items []DataItem
}

// DataRequest содержит запрос конкретного элемента данных по ID.
// Используется для получения детальной информации об элементе.
type DataRequest struct {
	ItemID string
}

// DataResponse содержит запрошенный элемент данных.
// Возвращается в ответ на DataRequest.
type DataResponse struct {
	Item DataItem
}

// SaveDataRequest содержит новый элемент данных для сохранения на сервере.
// Используется при создании новых записей.
type SaveDataRequest struct {
	Item NewDataItem
}

// SaveDataResponse содержит результат операции сохранения данных.
// Включает статус успеха и ID созданного элемента (если успешно).
type SaveDataResponse struct {
	Success bool
	Message string
	ItemID  string
}

// ErrorResponse содержит информацию об ошибке, произошедшей при обработке запроса.
// Используется для передачи деталей ошибки клиенту.
type ErrorResponse struct {
	Code    uint16
	Message string
}

// DeleteDataRequest содержит запрос на удаление элемента данных.
// Используется для удаления существующих записей.
type DeleteDataRequest struct {
	ItemID string
}

// DeleteDataResponse содержит результат операции удаления данных.
// Включает статус успеха и информационное сообщение.
type DeleteDataResponse struct {
	Success bool
	Message string
}

// UpdateDataRequest содержит запрос на обновление существующего элемента данных.
// Используется для модификации записей.
type UpdateDataRequest struct {
	ItemID string
	Item   NewDataItem
}

// UpdateDataResponse содержит результат операции обновления данных.
// Включает статус успеха и информационное сообщение.
type UpdateDataResponse struct {
	Success bool
	Message string
}

// DownloadRequest содержит запрос на загрузку данных элемента.
// Используется для получения бинарных данных.
type DownloadRequest struct {
	ItemID string
}

// DownloadResponse содержит данные запрошенного элемента.
// Используется для передачи бинарных данных клиенту.
type DownloadResponse struct {
	Success bool
	Data    []byte
	Message string
}
