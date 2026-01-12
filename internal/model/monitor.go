package model

import "time"

type InterfaceTraffic struct {
	Name string

	RxPacketsPerSecond   string
	RxBitsPerSecond      string
	FpRxPacketsPerSecond string
	FpRxBitsPerSecond    string

	RxDropsPerSecond  string
	RxErrorsPerSecond string

	TxPacketsPerSecond   string
	TxBitsPerSecond      string
	FpTxPacketsPerSecond string
	FpTxBitsPerSecond    string

	TxDropsPerSecond      string
	TxQueueDropsPerSecond string
	TxErrorsPerSecond     string

	Section string
}

type CustomerTrafficData struct {
	CustomerID         string      `json:"customer_id"`
	CustomerName       string      `json:"customer_name"`
	Username           string      `json:"username"`
	ServiceType        ServiceType `json:"service_type"`
	InterfaceName      string      `json:"interface_name"`
	RxBitsPerSecond    string      `json:"rx_bits_per_second"`
	TxBitsPerSecond    string      `json:"tx_bits_per_second"`
	RxPacketsPerSecond string      `json:"rx_packets_per_second"`
	TxPacketsPerSecond string      `json:"tx_packets_per_second"`
	DownloadSpeed      string      `json:"download_speed"`
	UploadSpeed        string      `json:"upload_speed"`
	Timestamp          time.Time   `json:"timestamp"`
}

type PingResponse struct {
	Seq        string `json:"seq"`
	Host       string `json:"host"`
	Size       string `json:"size"`
	TTL        string `json:"ttl"`
	Time       string `json:"time"`
	Status     string `json:"status"`
	Sent       string `json:"sent"`
	Received   string `json:"received"`
	PacketLoss string `json:"packet_loss"`
	AvgRtt     string `json:"avg_rtt"`
	MinRtt     string `json:"min_rtt"`
	MaxRtt     string `json:"max_rtt"`
	IsSummary  bool   `json:"is_summary"`
}
