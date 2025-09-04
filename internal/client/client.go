package client

import (
	"encoding/json"
	"fmt"
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

func NewClient(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	c.conn = conn
	return nil
}

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
		return nil, fmt.Errorf("failed to send message: %v", err)
	}

	buffer := make([]byte, 4096)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if n < 10 {
		return nil, fmt.Errorf("response too short: %d bytes", n)
	}

	header, payload, err := protocol.DeserializeMessage(buffer[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if header.Type == protocol.MsgTypeError {
		errorResp, err := protocol.DeserializeErrorResponse(payload)
		if err != nil {
			return nil, fmt.Errorf("error response: failed to parse")
		}
		return nil, fmt.Errorf("server error: %s", errorResp.Message)
	}

	return payload, nil
}

func (c *Client) Register(username, password string) error {
	req := protocol.RegisterRequest{
		Username: username,
		Password: password,
	}

	data, err := protocol.SerializeRegisterRequest(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %v", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeRegisterRequest, data)
	if err != nil {
		return err
	}

	resp, err := protocol.DeserializeRegisterResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	return nil
}

func (c *Client) Login(username, password string) error {
	req := protocol.AuthRequest{
		Username: username,
		Password: password,
	}

	data, err := protocol.SerializeAuthRequest(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %v", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeAuthRequest, data)
	if err != nil {
		return err
	}

	resp, err := protocol.DeserializeAuthResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("authentication failed")
	}

	c.token = resp.Token
	c.username = username
	return nil
}

func (c *Client) SyncData(lastSync time.Time) ([]protocol.DataItem, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	req := protocol.SyncRequest{
		LastSync: lastSync,
	}

	data, err := protocol.SerializeSyncRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %v", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeSyncRequest, data)
	if err != nil {
		return nil, err
	}

	resp, err := protocol.DeserializeSyncResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return resp.Items, nil
}

func (c *Client) GetData(itemID string) (protocol.DataItem, error) {
	if !c.IsAuthenticated() {
		return protocol.DataItem{}, fmt.Errorf("not authenticated")
	}

	req := protocol.DataRequest{
		ItemID: itemID,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return protocol.DataItem{}, fmt.Errorf("failed to serialize request: %v", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeDataRequest, data)
	if err != nil {
		return protocol.DataItem{}, err
	}

	var resp protocol.DataResponse
	if err := json.Unmarshal(response, &resp); err != nil {
		return protocol.DataItem{}, fmt.Errorf("failed to parse response: %v", err)
	}

	return resp.Item, nil
}

func (c *Client) SaveData(item protocol.NewDataItem) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	req := protocol.SaveDataRequest{
		Item: item,
	}

	data, err := protocol.SerializeSaveDataRequest(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %v", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeSaveDataRequest, data)
	if err != nil {
		return err
	}

	resp, err := protocol.DeserializeSaveDataResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to save data: %s", resp.Message)
	}

	return nil
}

func (c *Client) IsAuthenticated() bool {
	return c.token != "" && c.username != ""
}

func (c *Client) GetUsername() string {
	return c.username
}

func (c *Client) DeleteData(itemID string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	req := protocol.DeleteDataRequest{
		ItemID: itemID,
	}

	data, err := protocol.SerializeDeleteDataRequest(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %v", err)
	}

	response, err := c.sendAndReceive(protocol.MsgTypeDeleteDataRequest, data)
	if err != nil {
		return err
	}

	resp, err := protocol.DeserializeDeleteDataResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to delete data: %s", resp.Message)
	}

	return nil
}
