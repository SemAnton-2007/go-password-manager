package protocol

import (
	"encoding/binary"
	"encoding/json"
	"time"
)

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

	if len(data) < 10+int(header.Length) {
		return MessageHeader{}, nil, ErrInvalidMessage
	}

	return header, data[10 : 10+header.Length], nil
}

func SerializeAuthRequest(req AuthRequest) ([]byte, error) {
	return json.Marshal(req)
}

func DeserializeAuthRequest(data []byte) (AuthRequest, error) {
	var req AuthRequest
	err := json.Unmarshal(data, &req)
	return req, err
}

func SerializeAuthResponse(resp AuthResponse) ([]byte, error) {
	return json.Marshal(resp)
}

func DeserializeAuthResponse(data []byte) (AuthResponse, error) {
	var resp AuthResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

func SerializeRegisterRequest(req RegisterRequest) ([]byte, error) {
	return json.Marshal(req)
}

func DeserializeRegisterRequest(data []byte) (RegisterRequest, error) {
	var req RegisterRequest
	err := json.Unmarshal(data, &req)
	return req, err
}

func SerializeRegisterResponse(resp RegisterResponse) ([]byte, error) {
	return json.Marshal(resp)
}

func DeserializeRegisterResponse(data []byte) (RegisterResponse, error) {
	var resp RegisterResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

func SerializeSyncRequest(req SyncRequest) ([]byte, error) {
	return json.Marshal(struct {
		LastSync int64 `json:"last_sync"`
	}{
		LastSync: req.LastSync.UnixNano(),
	})
}

func DeserializeSyncRequest(data []byte) (SyncRequest, error) {
	var temp struct {
		LastSync int64 `json:"last_sync"`
	}
	err := json.Unmarshal(data, &temp)
	return SyncRequest{LastSync: time.Unix(0, temp.LastSync)}, err
}

func SerializeSyncResponse(resp SyncResponse) ([]byte, error) {
	return json.Marshal(resp)
}

func DeserializeSyncResponse(data []byte) (SyncResponse, error) {
	var resp SyncResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

func SerializeDataItem(item DataItem) ([]byte, error) {
	type dataItem struct {
		ID        string            `json:"id"`
		Type      uint8             `json:"type"`
		Name      string            `json:"name"`
		Data      []byte            `json:"data"`
		Metadata  map[string]string `json:"metadata"`
		CreatedAt int64             `json:"created_at"`
		UpdatedAt int64             `json:"updated_at"`
	}

	temp := dataItem{
		ID:        item.ID,
		Type:      item.Type,
		Name:      item.Name,
		Data:      item.Data,
		Metadata:  item.Metadata,
		CreatedAt: item.CreatedAt.UnixNano(),
		UpdatedAt: item.UpdatedAt.UnixNano(),
	}

	return json.Marshal(temp)
}

func DeserializeDataItem(data []byte) (DataItem, error) {
	type dataItem struct {
		ID        string            `json:"id"`
		Type      uint8             `json:"type"`
		Name      string            `json:"name"`
		Data      []byte            `json:"data"`
		Metadata  map[string]string `json:"metadata"`
		CreatedAt int64             `json:"created_at"`
		UpdatedAt int64             `json:"updated_at"`
	}

	var temp dataItem
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return DataItem{}, err
	}

	return DataItem{
		ID:        temp.ID,
		Type:      temp.Type,
		Name:      temp.Name,
		Data:      temp.Data,
		Metadata:  temp.Metadata,
		CreatedAt: time.Unix(0, temp.CreatedAt),
		UpdatedAt: time.Unix(0, temp.UpdatedAt),
	}, nil
}

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

func SerializeSaveDataResponse(resp SaveDataResponse) ([]byte, error) {
	return json.Marshal(resp)
}

func DeserializeSaveDataResponse(data []byte) (SaveDataResponse, error) {
	var resp SaveDataResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

func SerializeErrorResponse(resp ErrorResponse) ([]byte, error) {
	return json.Marshal(resp)
}

func DeserializeErrorResponse(data []byte) (ErrorResponse, error) {
	var resp ErrorResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

func SerializeDeleteDataRequest(req DeleteDataRequest) ([]byte, error) {
	return json.Marshal(req)
}

func DeserializeDeleteDataRequest(data []byte) (DeleteDataRequest, error) {
	var req DeleteDataRequest
	err := json.Unmarshal(data, &req)
	return req, err
}

func SerializeDeleteDataResponse(resp DeleteDataResponse) ([]byte, error) {
	return json.Marshal(resp)
}

func DeserializeDeleteDataResponse(data []byte) (DeleteDataResponse, error) {
	var resp DeleteDataResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}

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

func SerializeUpdateDataResponse(resp UpdateDataResponse) ([]byte, error) {
	return json.Marshal(resp)
}

func DeserializeUpdateDataResponse(data []byte) (UpdateDataResponse, error) {
	var resp UpdateDataResponse
	err := json.Unmarshal(data, &resp)
	return resp, err
}
