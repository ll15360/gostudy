package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

type dbConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
}

type appConfig struct {
	Database dbConfig `yaml:"database"`
}

func InitDB() {
	data, err := os.ReadFile("trpc_go.yaml")
	if err != nil {
		log.Fatalf("读取配置文件 trpc_go.yaml 失败: %v", err)
	}

	var cfg appConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	dbCfg := cfg.Database
	if dbCfg.User == "" {
		log.Fatal("数据库配置缺失：请在 trpc_go.yaml 中填写 database 段")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v\n请检查数据库配置与网络连接。", err)
	}

	log.Println("数据库连接成功，开始自动迁移表结构...")
	if err = DB.AutoMigrate(&PodInfo{}); err != nil {
		log.Fatalf("数据库表自动迁移失败: %v", err)
	}
	log.Println("数据库表自动迁移完成。")
}
