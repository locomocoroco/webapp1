package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

const userPepperPw = "4jhjj767o1ngl6dq"

var (
	ErrNotFound  = errors.New("resource not found")
	ErrInvalidID = errors.New("invalid id given")
	ErrInvalidPW = errors.New("invalid password provided")
)

type Users struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
}

func (us *UserService) ByID(id uint) (*Users, error) {
	var user Users
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

func (us *UserService) ByEmail(email string) (*Users, error) {
	var user Users
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err

}
func (us *UserService) Auth(email, password string) (*Users, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, ErrNotFound
	}
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+userPepperPw))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrInvalidPW
		default:
			return nil, err
		}
	}
	return foundUser, nil
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}
func (us *UserService) Update(user *Users) error {
	return us.db.Save(user).Error
}
func (us *UserService) Create(user *Users) error {
	pwBytes := []byte(user.Password + userPepperPw)
	passwordHash, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(passwordHash)
	user.Password = ""
	return us.db.Create(user).Error
}
func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := Users{Model: gorm.Model{ID: id}}
	return us.db.Delete(&user).Error
}

//NewUserService
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	return &UserService{
		db: db,
	}, nil
}

//Closes db conn
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&Users{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()
}
func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&Users{}).Error; err != nil {
		return err
	}
	return nil
}
func (us *UserService) Close() error {
	return us.db.Close()
}

type UserService struct {
	db *gorm.DB
}