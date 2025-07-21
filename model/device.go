package model

// DeviceInfo 代表设备信息
type DeviceInfo struct {
	IP        string `json:"ip"`
	Device    string `json:"device"`
	OS        string `json:"os"`
	Browser   string `json:"browser"`
	Timestamp int64  `json:"timestamp"`
}

// ClientInfo 代表一个客户端的完整信息
type ClientInfo struct {
	ID         string     `json:"id"`
	DeviceInfo DeviceInfo `json:"deviceInfo"`
}
