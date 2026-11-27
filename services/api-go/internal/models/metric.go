package models

import ("time")

type Metric struct {
	 ID          string    `json:"id" gorm:"primaryKey"`
    CPUUsage    float64   `json:"cpu_usage"`
    MemoryUsage float64   `json:"memory_usage"`
    DiskUsage   float64   `json:"disk_usage"`
    NetworkIn   float64   `json:"network_in"`
    NetworkOut  float64   `json:"network_out"`
    Timestamp   int64     `json:"timestamp"`
    CreatedAt   time.Time `json:"created_at"`
}