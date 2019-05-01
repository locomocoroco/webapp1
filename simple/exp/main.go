package main

import (
	"bufio"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"os"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "dbpass"
	dbname   = "simpleapes_dev"
)

type Users struct {
	gorm.Model
	Name   string
	Email  string `gorm:"not null;unique_index"`
	Colour string
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := gorm.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.DB().Ping(); err != nil {
		panic(err)
	}
	db.LogMode(true)
	db.AutoMigrate(&Users{})

	name, email, colour := getInfo()
	db.Create(&Users{Name: name, Email: email, Colour: colour})

	//db.DropTableIfExists(&User{})
}

func getInfo() (name, email, colour string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Whats your name?")
	name, _ = reader.ReadString('\n')
	fmt.Println("Whats your email?")
	email, _ = reader.ReadString('\n')
	fmt.Println("Whats your colour?")
	colour, _ = reader.ReadString('\n')
	return
}
