// pkg/genieacs/client.go
package genieacs

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/url"
// 	"time"

// 	genieacs_dto "mikrobill/pkg/genieacs/dto"
// )

// // Client adalah HTTP client untuk berkomunikasi dengan GenieACS API
// type Client struct {
// 	baseURL    string
// 	username   string
// 	password   string
// 	httpClient *http.Client
// }

// // Config untuk inisialisasi client
// type Config struct {
// 	BaseURL  string
// 	Username string
// 	Password string
// 	Timeout  time.Duration
// }

// // NewClient membuat instance baru GenieACS client
// func NewClient(cfg Config) *Client {
// 	if cfg.Timeout == 0 {
// 		cfg.Timeout = 10 * time.Second
// 	}

// 	return &Client{
// 		baseURL:  cfg.BaseURL,
// 		username: cfg.Username,
// 		password: cfg.Password,
// 		httpClient: &http.Client{
// 			Timeout: cfg.Timeout,
// 		},
// 	}
// }

// // doRequest melakukan HTTP request ke GenieACS API
// func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
// 	var reqBody io.Reader
// 	if body != nil {
// 		jsonData, err := json.Marshal(body)
// 		if err != nil {
// 			return nil, fmt.Errorf("marshal request body: %w", err)
// 		}
// 		reqBody = bytes.NewBuffer(jsonData)
// 	}

// 	fullURL := c.baseURL + path
// 	req, err := http.NewRequest(method, fullURL, reqBody)
// 	if err != nil {
// 		return nil, fmt.Errorf("create request: %w", err)
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Accept", "application/json")

// 	if c.username != "" && c.password != "" {
// 		req.SetBasicAuth(c.username, c.password)
// 	}

// 	resp, err := c.httpClient.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("execute request: %w", err)
// 	}

// 	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
// 		bodyBytes, _ := io.ReadAll(resp.Body)
// 		resp.Body.Close()
// 		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
// 	}

// 	return resp, nil
// }

// // GetDevices mengambil daftar semua devices
// func (c *Client) GetDevices(query map[string]interface{}) ([]Device, error) {
// 	path := "/devices/"

// 	if len(query) > 0 {
// 		queryJSON, err := json.Marshal(query)
// 		if err != nil {
// 			return nil, fmt.Errorf("marshal query: %w", err)
// 		}
// 		path += "?query=" + url.QueryEscape(string(queryJSON))
// 	}

// 	resp, err := c.doRequest("GET", path, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var devices []Device
// 	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
// 		return nil, fmt.Errorf("decode response: %w", err)
// 	}

// 	return devices, nil
// }

// // GetDevice mengambil detail device berdasarkan ID
// func (c *Client) GetDevice(deviceID string) (*Device, error) {
// 	path := "/devices/" + url.PathEscape(deviceID)

// 	resp, err := c.doRequest("GET", path, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var device Device
// 	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
// 		return nil, fmt.Errorf("decode response: %w", err)
// 	}

// 	return &device, nil
// }

// // FindDeviceByPhoneNumber mencari device berdasarkan tag nomor telepon
// func (c *Client) FindDeviceByPhoneNumber(phoneNumber string) (*Device, error) {
// 	query := map[string]interface{}{
// 		"_tags": phoneNumber,
// 	}

// 	devices, err := c.GetDevices(query)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(devices) == 0 {
// 		return nil, fmt.Errorf("no device found with phone number: %s", phoneNumber)
// 	}

// 	return &devices[0], nil
// }

// // FindDeviceByPPPoE mencari device berdasarkan PPPoE username
// func (c *Client) FindDeviceByPPPoE(username string) (*Device, error) {
// 	// Coba beberapa path yang mungkin berisi PPPoE username
// 	paths := []string{
// 		"InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.Username",
// 		"VirtualParameters.pppoeUsername",
// 	}

// 	orConditions := make([]map[string]interface{}, len(paths))
// 	for i, path := range paths {
// 		orConditions[i] = map[string]interface{}{
// 			path: username,
// 		}
// 	}

// 	query := map[string]interface{}{
// 		"$or": orConditions,
// 	}

// 	devices, err := c.GetDevices(query)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(devices) == 0 {
// 		return nil, fmt.Errorf("no device found with PPPoE username: %s", username)
// 	}

// 	return &devices[0], nil
// }

// // SetParameterValues mengatur nilai parameter pada device
// func (c *Client) SetParameterValues(req *genieacs_model.SetParameterValuesRequest) (*genieacs_model.TaskResponse, error) {
// 	path := fmt.Sprintf("/devices/%s/tasks", url.PathEscape(req.DeviceID))

// 	if req.ConnectionRequest {
// 		path += "?connection_request"
// 	}

// 	task := map[string]interface{}{
// 		"name":            "setParameterValues",
// 		"parameterValues": req.ParameterValues,
// 	}

// 	resp, err := c.doRequest("POST", path, task)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var taskResp genieacs_model.TaskResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
// 		return nil, fmt.Errorf("decode response: %w", err)
// 	}

// 	return &taskResp, nil
// }

// // Reboot melakukan reboot pada device
// func (c *Client) Reboot(deviceID string) (*genieacs_model.TaskResponse, error) {
// 	path := fmt.Sprintf("/devices/%s/tasks", url.PathEscape(deviceID))

// 	task := map[string]interface{}{
// 		"name":      "reboot",
// 		"timestamp": time.Now().Format(time.RFC3339),
// 	}

// 	resp, err := c.doRequest("POST", path, task)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var taskResp genieacs_model.TaskResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
// 		return nil, fmt.Errorf("decode response: %w", err)
// 	}

// 	return &taskResp, nil
// }

// // FactoryReset melakukan factory reset pada device
// func (c *Client) FactoryReset(deviceID string) (*genieacs_model.TaskResponse, error) {
// 	path := fmt.Sprintf("/devices/%s/tasks", url.PathEscape(deviceID))

// 	task := map[string]interface{}{
// 		"name":      "factoryReset",
// 		"timestamp": time.Now().Format(time.RFC3339),
// 	}

// 	resp, err := c.doRequest("POST", path, task)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var taskResp genieacs_model.TaskResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
// 		return nil, fmt.Errorf("decode response: %w", err)
// 	}

// 	return &taskResp, nil
// }

// // RefreshObject melakukan refresh pada object device
// func (c *Client) RefreshObject(req *genieacs_model.RefreshObjectRequest) (*genieacs_model.TaskResponse, error) {
// 	path := fmt.Sprintf("/devices/%s/tasks", url.PathEscape(req.DeviceID))

// 	if req.ConnectionRequest {
// 		path += "?connection_request"
// 	}

// 	task := map[string]interface{}{
// 		"name":       "refreshObject",
// 		"objectName": req.ObjectName,
// 	}

// 	resp, err := c.doRequest("POST", path, task)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var taskResp genieacs_model.TaskResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
// 		return nil, fmt.Errorf("decode response: %w", err)
// 	}

// 	return &taskResp, nil
// }

// // AddTag menambahkan tag pada device
// func (c *Client) AddTag(deviceID, tag string) error {
// 	// Get device terlebih dahulu untuk mendapatkan tags yang ada
// 	device, err := c.GetDevice(deviceID)
// 	if err != nil {
// 		return err
// 	}

// 	// Cek apakah tag sudah ada
// 	for _, t := range device.Tags {
// 		if t == tag {
// 			return nil // Tag sudah ada
// 		}
// 	}

// 	// Tambahkan tag baru
// 	newTags := append(device.Tags, tag)

// 	// Update device
// 	path := "/devices/" + url.PathEscape(deviceID)
// 	updateBody := map[string]interface{}{
// 		"_tags": newTags,
// 	}

// 	resp, err := c.doRequest("PUT", path, updateBody)
// 	if err != nil {
// 		return err
// 	}
// 	resp.Body.Close()

// 	return nil
// }

// // RemoveTag menghapus tag dari device
// func (c *Client) RemoveTag(deviceID, tag string) error {
// 	// Get device terlebih dahulu
// 	device, err := c.GetDevice(deviceID)
// 	if err != nil {
// 		return err
// 	}

// 	// Filter tag yang akan dihapus
// 	newTags := make([]string, 0)
// 	found := false
// 	for _, t := range device.Tags {
// 		if t != tag {
// 			newTags = append(newTags, t)
// 		} else {
// 			found = true
// 		}
// 	}

// 	if !found {
// 		return nil // Tag tidak ada
// 	}

// 	// Update device
// 	path := "/devices/" + url.PathEscape(deviceID)
// 	updateBody := map[string]interface{}{
// 		"_tags": newTags,
// 	}

// 	resp, err := c.doRequest("PUT", path, updateBody)
// 	if err != nil {
// 		return err
// 	}
// 	resp.Body.Close()

// 	return nil
// }

// // MonitorRXPower memonitor RX Power dari semua devices dan mengembalikan device dengan RX Power di bawah threshold
// func (c *Client) MonitorRXPower(threshold float64) ([]CriticalDevice, error) {
// 	devices, err := c.GetDevices(nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var criticalDevices []CriticalDevice

// 	for _, device := range devices {
// 		rxPower := device.GetRXPower()
// 		if rxPower != nil && *rxPower < threshold {
// 			criticalDevices = append(criticalDevices, CriticalDevice{
// 				DeviceID:     device.ID,
// 				SerialNumber: device.GetSerialNumber(),
// 				RXPower:      *rxPower,
// 				LastInform:   device.LastInform,
// 				PPPoEUsername: device.GetPPPoEUsername(),
// 			})
// 		}
// 	}

// 	return criticalDevices, nil
// }

// // MonitorOfflineDevices memonitor devices yang offline lebih dari threshold (dalam jam)
// func (c *Client) MonitorOfflineDevices(thresholdHours int) ([]OfflineDevice, error) {
// 	devices, err := c.GetDevices(nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var offlineDevices []OfflineDevice
// 	now := time.Now()
// 	thresholdDuration := time.Duration(thresholdHours) * time.Hour

// 	for _, device := range devices {
// 		if device.LastInform == nil {
// 			continue
// 		}

// 		lastInform, err := time.Parse(time.RFC3339, *device.LastInform)
// 		if err != nil {
// 			continue
// 		}

// 		timeSinceInform := now.Sub(lastInform)
// 		if timeSinceInform > thresholdDuration {
// 			offlineDevices = append(offlineDevices, OfflineDevice{
// 				DeviceID:      device.ID,
// 				SerialNumber:  device.GetSerialNumber(),
// 				PPPoEUsername: device.GetPPPoEUsername(),
// 				LastInform:    *device.LastInform,
// 				OfflineHours:  timeSinceInform.Hours(),
// 			})
// 		}
// 	}

// 	return offlineDevices, nil
// }
