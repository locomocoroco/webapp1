package main

import (
	"encoding/json"
	"fmt"
	"os"
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
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACKey  string         `json:"hmac_key"`
	Database PostfresConfig `json:"database"`
	Mailgun  MailgunConfig  `json:"mailgun"`
}

func (c ConfigV) IsProd() bool {
	return c.Env == "prod"
}
func DefaultConfig() ConfigV {
	return ConfigV{
		Port:     3000,
		Env:      "dev",
		Pepper:   "4jhjj767o1ngl6dq",
		HMACKey:  "5gfl7lhl76lle7gh",
		Database: DefaultPostgresConfig(),
	}
}

type MailgunConfig struct {
	APIKey       string `json:"api_key"`
	PublicAPIKEY string `json:"public_api_key"`
	Domain       string `json:"domain"`
}

func LoadConfig(prod bool) ConfigV {
	f, err := os.Open(".config")
	if err != nil {
		if prod {
			panic(err)
		}
		fmt.Println("using default config")
		return DefaultConfig()
	}
	var c ConfigV
	dec := json.NewDecoder(f)
	err = dec.Decode(&c)
	if err != nil {
		panic(err)
	}
	fmt.Println("loaded passed config")
	return c
}

//db, err := gorm.Open("postgres", connectionInfo)
//if err != nil {
//return nil, err
//}
//db.LogMode(true)
