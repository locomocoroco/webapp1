package main

import (
	"fmt"
)

type PostfresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (c PostfresConfig) Dialect() string {
	return "postgres"
}
func (c PostfresConfig) ConnectionInfo() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name)
}
func DefaultPostgresConfig() PostfresConfig {
	return PostfresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "dbpass",
		Name:     "simpleapes_dev",
	}
}

type ConfigV struct {
	Port    int    `json:"port"`
	Env     string `json:"env"`
	Pepper  string `json:"pepper"`
	HMACKey string `json:"hmac_key"`
}

func (c ConfigV) IsProd() bool {
	return c.Env == "prod"
}
func DefaultConfig() ConfigV {
	return ConfigV{
		Port:    3000,
		Env:     "dev",
		Pepper:  "4jhjj767o1ngl6dq",
		HMACKey: "5gfl7lhl76lle7gh",
	}
}

//db, err := gorm.Open("postgres", connectionInfo)
//if err != nil {
//return nil, err
//}
//db.LogMode(true)
