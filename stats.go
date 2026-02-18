package solarman

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Station struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	GenerationPower    float64 `json:"generationPower"`
	InstalledCapacity  float64 `json:"installedCapacity"`
	NetworkStatus      string  `json:"networkStatus"`
	LocationAddress    string  `json:"locationAddress"`
	GridInterconnType  string  `json:"gridInterconnectionType"`
	LastUpdateTime     int     `json:"lastUpdateTime"`
	StartOperatingTime int     `json:"startOperatingTime"`
}

type Device struct {
	DeviceID       int    `json:"deviceId"`
	DeviceSn       string `json:"deviceSn"`
	DeviceType     string `json:"deviceType"`
	ConnectStatus  int    `json:"connectStatus"`
	CollectionTime int    `json:"collectionTime"`
}

func (c *Client) Stations() ([]Station, error) {
	bts, err := c.post(baseURL+"/station/v1.0/list", "{}")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Success     bool      `json:"success"`
		Msg         any       `json:"msg"`
		StationList []Station `json:"stationList"`
	}
	if err := json.Unmarshal(bts, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal stations: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("stations: %v", resp.Msg)
	}
	return resp.StationList, nil
}

func (c *Client) StationDevices(stationID int) ([]Device, error) {
	body := fmt.Sprintf(`{"stationId":%d}`, stationID)
	bts, err := c.post(baseURL+"/station/v1.0/device", body)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Success         bool     `json:"success"`
		Msg             any      `json:"msg"`
		DeviceListItems []Device `json:"deviceListItems"`
	}
	if err := json.Unmarshal(bts, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal devices: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("station devices: %v", resp.Msg)
	}
	return resp.DeviceListItems, nil
}

func (c *Client) CurrentData(deviceID int) (CurrentData, error) {
	body := fmt.Sprintf(`{"deviceId":%d}`, deviceID)
	bts, err := c.post(
		fmt.Sprintf(baseURL+"/device/v1.0/currentData?appId=%s&language=en", c.appID),
		body,
	)
	if err != nil {
		return CurrentData{}, err
	}

	var data CurrentData
	if err := json.Unmarshal(bts, &data); err != nil {
		return CurrentData{}, fmt.Errorf("unmarshal currentData: %w", err)
	}
	return data, nil
}

func (c *Client) post(url, body string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	return io.ReadAll(resp.Body)
}

type CurrentData struct {
	Code           any        `json:"code"`
	Msg            any        `json:"msg"`
	Success        bool       `json:"success"`
	RequestID      string     `json:"requestId"`
	DeviceSn       string     `json:"deviceSn"`
	DeviceID       int        `json:"deviceId"`
	DeviceType     string     `json:"deviceType"`
	DeviceState    int        `json:"deviceState"`
	CollectionTime int        `json:"collectionTime"`
	DataList       []DataList `json:"dataList"`
}

type DataList struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Unit  any    `json:"unit"`
	Name  string `json:"name"`
}
