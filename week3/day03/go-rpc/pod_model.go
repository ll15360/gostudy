package main

import (
	"time"
)

type PodInfo struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;comment:主键ID"`
	PodName       string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_pod_namespace;comment:Pod 名称"`
	Namespace     string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_pod_namespace;comment:命名空间"`
	Image         string    `gorm:"type:varchar(255);not null;comment:容器镜像"`
	Status        string    `gorm:"type:varchar(50);not null;comment:状态"`
	Labels        string    `gorm:"type:json;comment:标签信息"`
	CPURequest    string    `gorm:"type:varchar(50);default:'';column:cpu_request;comment:CPU 请求配额"`
	CPULimit      string    `gorm:"type:varchar(50);default:'';column:cpu_limit;comment:CPU 限制配额"`
	MemoryRequest string    `gorm:"type:varchar(50);default:'';column:memory_request;comment:内存请求配额"`
	MemoryLimit   string    `gorm:"type:varchar(50);default:'';column:memory_limit;comment:内存限制配额"`
	CreatedAt     time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt     time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间"`
}

func (PodInfo) TableName() string {
	return "pods_info"
}
