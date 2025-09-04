package protocol

import (
	"errors"
	"time"
)

// Типы сообщений
const (
	MsgTypeAuthRequest        = 0x01
	MsgTypeAuthResponse       = 0x02
	MsgTypeRegisterRequest    = 0x03
	MsgTypeRegisterResponse   = 0x04
	MsgTypeSyncRequest        = 0x05
	MsgTypeSyncResponse       = 0x06
	MsgTypeDataRequest        = 0x07
	MsgTypeDataResponse       = 0x08
	MsgTypeSaveDataRequest    = 0x09
	MsgTypeSaveDataResponse   = 0x0A
	MsgTypeError              = 0xFF
	MsgTypeDeleteDataRequest  = 0x0B
	MsgTypeDeleteDataResponse = 0x0C
	MsgTypeUpdateDataRequest  = 0x0D
	MsgTypeUpdateDataResponse = 0x0E
)

// Типы данных
const (
	DataTypeLoginPassword = 0x01
	DataTypeText          = 0x02
	DataTypeBinary        = 0x03
	DataTypeBankCard      = 0x04
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrAuthFailed     = errors.New("authentication failed")
)

// Базовая структура сообщения
type MessageHeader struct {
	Type      uint8
	Version   uint8
	MessageID uint32
	Length    uint32
}

// Структуры для аутентификации
type AuthRequest struct {
	Username string
	Password string
}

type AuthResponse struct {
	Success bool
	Token   string
}

// Структуры для регистрации
type RegisterRequest struct {
	Username string
	Password string
}

type RegisterResponse struct {
	Success bool
	Message string
}

// Структуры данных
type DataItem struct {
	ID        string
	Type      uint8
	Name      string
	Data      []byte
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Структура для создания нового элемента
type NewDataItem struct {
	Type     uint8
	Name     string
	Data     []byte
	Metadata map[string]string
}

type SyncRequest struct {
	LastSync time.Time
}

type SyncResponse struct {
	Items []DataItem
}

type DataRequest struct {
	ItemID string
}

type DataResponse struct {
	Item DataItem
}

type SaveDataRequest struct {
	Item NewDataItem
}

type SaveDataResponse struct {
	Success bool
	Message string
	ItemID  string
}

type ErrorResponse struct {
	Code    uint16
	Message string
}

type DeleteDataRequest struct {
	ItemID string
}

type DeleteDataResponse struct {
	Success bool
	Message string
}

type UpdateDataRequest struct {
	ItemID string
	Item   NewDataItem
}

type UpdateDataResponse struct {
	Success bool
	Message string
}
