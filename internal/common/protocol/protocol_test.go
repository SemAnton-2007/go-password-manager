package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSerializeDeserializeMessage(t *testing.T) {
	testData := []byte("test message data")

	// Сериализация
	message := SerializeMessage(MsgTypeAuthRequest, 123, testData)

	if len(message) < 10+len(testData) {
		t.Errorf("Message too short. Expected at least %d, got %d",
			10+len(testData), len(message))
	}

	// Десериализация
	header, payload, err := DeserializeMessage(message)
	if err != nil {
		t.Fatalf("DeserializeMessage failed: %v", err)
	}

	if header.Type != MsgTypeAuthRequest {
		t.Errorf("Expected type %d, got %d", MsgTypeAuthRequest, header.Type)
	}

	if header.MessageID != 123 {
		t.Errorf("Expected message ID 123, got %d", header.MessageID)
	}

	if header.Length != uint32(len(testData)) {
		t.Errorf("Expected length %d, got %d", len(testData), header.Length)
	}

	if string(payload) != string(testData) {
		t.Errorf("Payload mismatch. Got: %s, Expected: %s",
			string(payload), string(testData))
	}
}

func TestAuthRequestResponse(t *testing.T) {
	// AuthRequest
	authReq := AuthRequest{
		Username: "testuser",
		Password: "testpass",
	}

	data, err := SerializeAuthRequest(authReq)
	if err != nil {
		t.Fatalf("SerializeAuthRequest failed: %v", err)
	}

	authReq2, err := DeserializeAuthRequest(data)
	if err != nil {
		t.Fatalf("DeserializeAuthRequest failed: %v", err)
	}

	if authReq2.Username != authReq.Username {
		t.Errorf("Username mismatch. Got: %s, Expected: %s",
			authReq2.Username, authReq.Username)
	}

	// AuthResponse
	authResp := AuthResponse{
		Success: true,
		Token:   "testtoken",
	}

	data, err = SerializeAuthResponse(authResp)
	if err != nil {
		t.Fatalf("SerializeAuthResponse failed: %v", err)
	}

	authResp2, err := DeserializeAuthResponse(data)
	if err != nil {
		t.Fatalf("DeserializeAuthResponse failed: %v", err)
	}

	if authResp2.Success != authResp.Success || authResp2.Token != authResp.Token {
		t.Error("AuthResponse mismatch")
	}
}

func TestDataItemSerialization(t *testing.T) {
	now := time.Now()
	item := DataItem{
		ID:   "test-id",
		Type: DataTypeLoginPassword,
		Name: "Test Item",
		Data: []byte("test data"),
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := SerializeDataItem(item)
	if err != nil {
		t.Fatalf("SerializeDataItem failed: %v", err)
	}

	item2, err := DeserializeDataItem(data)
	if err != nil {
		t.Fatalf("DeserializeDataItem failed: %v", err)
	}

	if item2.ID != item.ID || item2.Type != item.Type || item2.Name != item.Name {
		t.Error("DataItem basic fields mismatch")
	}

	if string(item2.Data) != string(item.Data) {
		t.Error("DataItem data mismatch")
	}

	if len(item2.Metadata) != len(item.Metadata) {
		t.Errorf("Metadata length mismatch. Got: %d, Expected: %d",
			len(item2.Metadata), len(item.Metadata))
	}

	for k, v := range item.Metadata {
		if item2.Metadata[k] != v {
			t.Errorf("Metadata key %s mismatch. Got: %s, Expected: %s",
				k, item2.Metadata[k], v)
		}
	}
}

func TestSyncRequestResponse(t *testing.T) {
	// SyncRequest
	syncTime := time.Now().Add(-time.Hour)
	syncReq := SyncRequest{LastSync: syncTime}

	data, err := SerializeSyncRequest(syncReq)
	if err != nil {
		t.Fatalf("SerializeSyncRequest failed: %v", err)
	}

	syncReq2, err := DeserializeSyncRequest(data)
	if err != nil {
		t.Fatalf("DeserializeSyncRequest failed: %v", err)
	}

	if syncReq2.LastSync.Truncate(time.Second) != syncReq.LastSync.Truncate(time.Second) {
		t.Error("SyncRequest time mismatch")
	}

	// SyncResponse
	items := []DataItem{
		{ID: "1", Name: "Item1", Type: DataTypeText},
		{ID: "2", Name: "Item2", Type: DataTypeLoginPassword},
	}
	syncResp := SyncResponse{Items: items}

	data, err = SerializeSyncResponse(syncResp)
	if err != nil {
		t.Fatalf("SerializeSyncResponse failed: %v", err)
	}

	syncResp2, err := DeserializeSyncResponse(data)
	if err != nil {
		t.Fatalf("DeserializeSyncResponse failed: %v", err)
	}

	if len(syncResp2.Items) != len(syncResp.Items) {
		t.Errorf("Items length mismatch. Got: %d, Expected: %d",
			len(syncResp2.Items), len(syncResp.Items))
	}
}

func TestSaveDataRequest(t *testing.T) {
	item := NewDataItem{
		Type: DataTypeText,
		Name: "Test Item",
		Data: []byte("test data"),
		Metadata: map[string]string{
			"meta1": "value1",
		},
	}

	req := SaveDataRequest{Item: item}

	data, err := SerializeSaveDataRequest(req)
	if err != nil {
		t.Fatalf("SerializeSaveDataRequest failed: %v", err)
	}

	req2, err := DeserializeSaveDataRequest(data)
	if err != nil {
		t.Fatalf("DeserializeSaveDataRequest failed: %v", err)
	}

	if req2.Item.Type != req.Item.Type || req2.Item.Name != req.Item.Name {
		t.Error("SaveDataRequest basic fields mismatch")
	}

	if string(req2.Item.Data) != string(req.Item.Data) {
		t.Error("SaveDataRequest data mismatch")
	}

	if len(req2.Item.Metadata) != len(req.Item.Metadata) {
		t.Errorf("Metadata length mismatch. Got: %d, Expected: %d",
			len(req2.Item.Metadata), len(req.Item.Metadata))
	}
}

func TestErrorResponse(t *testing.T) {
	errorResp := ErrorResponse{
		Code:    500,
		Message: "Test error",
	}

	data, err := SerializeErrorResponse(errorResp)
	if err != nil {
		t.Fatalf("SerializeErrorResponse failed: %v", err)
	}

	errorResp2, err := DeserializeErrorResponse(data)
	if err != nil {
		t.Fatalf("DeserializeErrorResponse failed: %v", err)
	}

	if errorResp2.Code != errorResp.Code || errorResp2.Message != errorResp.Message {
		t.Error("ErrorResponse mismatch")
	}
}

func TestInvalidMessage(t *testing.T) {
	// Слишком короткое сообщение
	shortData := []byte{0x01, 0x01}
	_, _, err := DeserializeMessage(shortData)
	if err == nil {
		t.Error("Should fail with short message")
	}

	// Неверный JSON
	invalidJSON := []byte("{invalid json")
	header := MessageHeader{Length: uint32(len(invalidJSON))}
	message := SerializeMessage(MsgTypeAuthRequest, 1, invalidJSON)

	_, payload, _ := DeserializeMessage(message)
	if len(payload) != int(header.Length) {
		t.Error("Should return payload even for invalid JSON")
	}
}

func TestUpdateDataRequest(t *testing.T) {
	item := NewDataItem{
		Type: DataTypeBankCard,
		Name: "Credit Card",
		Data: []byte("card data"),
		Metadata: map[string]string{
			"bank": "test bank",
		},
	}

	req := UpdateDataRequest{
		ItemID: "test-id",
		Item:   item,
	}

	data, err := SerializeUpdateDataRequest(req)
	if err != nil {
		t.Fatalf("SerializeUpdateDataRequest failed: %v", err)
	}

	req2, err := DeserializeUpdateDataRequest(data)
	if err != nil {
		t.Fatalf("DeserializeUpdateDataRequest failed: %v", err)
	}

	if req2.ItemID != req.ItemID {
		t.Errorf("ItemID mismatch. Got: %s, Expected: %s", req2.ItemID, req.ItemID)
	}

	if req2.Item.Type != req.Item.Type {
		t.Error("Item type mismatch")
	}
}

func TestRegisterRequestResponse(t *testing.T) {
	// RegisterRequest
	registerReq := RegisterRequest{
		Username: "newuser",
		Password: "newpass",
	}

	data, err := SerializeRegisterRequest(registerReq)
	if err != nil {
		t.Fatalf("SerializeRegisterRequest failed: %v", err)
	}

	registerReq2, err := DeserializeRegisterRequest(data)
	if err != nil {
		t.Fatalf("DeserializeRegisterRequest failed: %v", err)
	}

	if registerReq2.Username != registerReq.Username {
		t.Errorf("Username mismatch. Got: %s, Expected: %s",
			registerReq2.Username, registerReq.Username)
	}

	// RegisterResponse
	registerResp := RegisterResponse{
		Success: true,
		Message: "User created",
	}

	data, err = SerializeRegisterResponse(registerResp)
	if err != nil {
		t.Fatalf("SerializeRegisterResponse failed: %v", err)
	}

	registerResp2, err := DeserializeRegisterResponse(data)
	if err != nil {
		t.Fatalf("DeserializeRegisterResponse failed: %v", err)
	}

	if registerResp2.Success != registerResp.Success || registerResp2.Message != registerResp.Message {
		t.Error("RegisterResponse mismatch")
	}
}

func TestSaveDataResponse(t *testing.T) {
	saveResp := SaveDataResponse{
		Success: true,
		Message: "Data saved",
		ItemID:  "test-id",
	}

	data, err := SerializeSaveDataResponse(saveResp)
	if err != nil {
		t.Fatalf("SerializeSaveDataResponse failed: %v", err)
	}

	saveResp2, err := DeserializeSaveDataResponse(data)
	if err != nil {
		t.Fatalf("DeserializeSaveDataResponse failed: %v", err)
	}

	if saveResp2.Success != saveResp.Success || saveResp2.Message != saveResp.Message || saveResp2.ItemID != saveResp.ItemID {
		t.Error("SaveDataResponse mismatch")
	}
}

func TestDeleteDataRequestResponse(t *testing.T) {
	// DeleteDataRequest
	deleteReq := DeleteDataRequest{
		ItemID: "test-id",
	}

	data, err := SerializeDeleteDataRequest(deleteReq)
	if err != nil {
		t.Fatalf("SerializeDeleteDataRequest failed: %v", err)
	}

	deleteReq2, err := DeserializeDeleteDataRequest(data)
	if err != nil {
		t.Fatalf("DeserializeDeleteDataRequest failed: %v", err)
	}

	if deleteReq2.ItemID != deleteReq.ItemID {
		t.Errorf("ItemID mismatch. Got: %s, Expected: %s", deleteReq2.ItemID, deleteReq.ItemID)
	}

	// DeleteDataResponse
	deleteResp := DeleteDataResponse{
		Success: true,
		Message: "Data deleted",
	}

	data, err = SerializeDeleteDataResponse(deleteResp)
	if err != nil {
		t.Fatalf("SerializeDeleteDataResponse failed: %v", err)
	}

	deleteResp2, err := DeserializeDeleteDataResponse(data)
	if err != nil {
		t.Fatalf("DeserializeDeleteDataResponse failed: %v", err)
	}

	if deleteResp2.Success != deleteResp.Success || deleteResp2.Message != deleteResp.Message {
		t.Error("DeleteDataResponse mismatch")
	}
}

func TestUpdateDataResponse(t *testing.T) {
	updateResp := UpdateDataResponse{
		Success: true,
		Message: "Data updated",
	}

	data, err := SerializeUpdateDataResponse(updateResp)
	if err != nil {
		t.Fatalf("SerializeUpdateDataResponse failed: %v", err)
	}

	updateResp2, err := DeserializeUpdateDataResponse(data)
	if err != nil {
		t.Fatalf("DeserializeUpdateDataResponse failed: %v", err)
	}

	if updateResp2.Success != updateResp.Success || updateResp2.Message != updateResp.Message {
		t.Error("UpdateDataResponse mismatch")
	}
}

func TestDownloadRequestResponse(t *testing.T) {
	// DownloadRequest
	downloadReq := DownloadRequest{
		ItemID: "test-id",
	}

	data, err := SerializeDownloadRequest(downloadReq)
	if err != nil {
		t.Fatalf("SerializeDownloadRequest failed: %v", err)
	}

	downloadReq2, err := DeserializeDownloadRequest(data)
	if err != nil {
		t.Fatalf("DeserializeDownloadRequest failed: %v", err)
	}

	if downloadReq2.ItemID != downloadReq.ItemID {
		t.Errorf("ItemID mismatch. Got: %s, Expected: %s", downloadReq2.ItemID, downloadReq.ItemID)
	}

	// DownloadResponse
	testData := []byte("test download data")
	downloadResp := DownloadResponse{
		Success: true,
		Data:    testData,
		Message: "Download successful",
	}

	data, err = SerializeDownloadResponse(downloadResp)
	if err != nil {
		t.Fatalf("SerializeDownloadResponse failed: %v", err)
	}

	downloadResp2, err := DeserializeDownloadResponse(data)
	if err != nil {
		t.Fatalf("DeserializeDownloadResponse failed: %v", err)
	}

	if downloadResp2.Success != downloadResp.Success ||
		string(downloadResp2.Data) != string(downloadResp.Data) ||
		downloadResp2.Message != downloadResp.Message {
		t.Error("DownloadResponse mismatch")
	}
}

func TestDeserializeHeader(t *testing.T) {
	testData := []byte("test data")
	message := SerializeMessage(MsgTypeAuthRequest, 123, testData)

	header, err := DeserializeHeader(message[:10])
	if err != nil {
		t.Fatalf("DeserializeHeader failed: %v", err)
	}

	if header.Type != MsgTypeAuthRequest {
		t.Errorf("Expected type %d, got %d", MsgTypeAuthRequest, header.Type)
	}

	if header.MessageID != 123 {
		t.Errorf("Expected message ID 123, got %d", header.MessageID)
	}

	if header.Length != uint32(len(testData)) {
		t.Errorf("Expected length %d, got %d", len(testData), header.Length)
	}

	// Тест с невалидным заголовком
	shortHeader := []byte{0x01, 0x01} // слишком короткий
	_, err = DeserializeHeader(shortHeader)
	if err == nil {
		t.Error("Should fail with short header")
	}
}

func TestDataRequestResponse(t *testing.T) {
	// DataRequest
	dataReq := DataRequest{
		ItemID: "test-id",
	}

	data, err := json.Marshal(dataReq)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var dataReq2 DataRequest
	err = json.Unmarshal(data, &dataReq2)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if dataReq2.ItemID != dataReq.ItemID {
		t.Errorf("ItemID mismatch. Got: %s, Expected: %s", dataReq2.ItemID, dataReq.ItemID)
	}

	// DataResponse
	now := time.Now()
	dataItem := DataItem{
		ID:   "test-id",
		Type: DataTypeText,
		Name: "Test Item",
		Data: []byte("test data"),
		Metadata: map[string]string{
			"key": "value",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	dataResp := DataResponse{
		Item: dataItem,
	}

	data, err = json.Marshal(dataResp)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var dataResp2 DataResponse
	err = json.Unmarshal(data, &dataResp2)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if dataResp2.Item.ID != dataResp.Item.ID {
		t.Errorf("Item ID mismatch. Got: %s, Expected: %s", dataResp2.Item.ID, dataResp.Item.ID)
	}
}

func TestMessageWithEmptyPayload(t *testing.T) {
	// Сообщение с пустым payload
	message := SerializeMessage(MsgTypeAuthRequest, 123, []byte{})

	header, payload, err := DeserializeMessage(message)
	if err != nil {
		t.Fatalf("DeserializeMessage failed: %v", err)
	}

	if header.Length != 0 {
		t.Errorf("Expected length 0, got %d", header.Length)
	}

	if len(payload) != 0 {
		t.Errorf("Expected empty payload, got %d bytes", len(payload))
	}
}

func TestEdgeCases(t *testing.T) {
	// Тестируем крайние случаи с большими данными
	largeData := make([]byte, 1024*10) // 10KB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	message := SerializeMessage(MsgTypeAuthRequest, 1, largeData)
	header, payload, err := DeserializeMessage(message)
	if err != nil {
		t.Fatalf("DeserializeMessage with large data failed: %v", err)
	}

	if header.Length != uint32(len(largeData)) {
		t.Errorf("Large data length mismatch. Got: %d, Expected: %d", header.Length, len(largeData))
	}

	if len(payload) != len(largeData) {
		t.Errorf("Large data payload length mismatch. Got: %d, Expected: %d", len(payload), len(largeData))
	}
}
