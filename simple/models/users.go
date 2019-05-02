package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
	"webapp1/simple/hash"
	"webapp1/simple/rand"
)

const userPepperPw = "4jhjj767o1ngl6dq"
const hmacSecretKey = "5gfl7lhl76lle7gh"

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
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
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
func (us *UserService) ByRemember(token string) (*Users, error) {
	var user Users
	rememberHash := us.hmac.Hash(token)
	err := first(us.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
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
	if user.Remember != "" {
		user.RememberHash = us.hmac.Hash(user.Remember)
	}
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
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
	}

	user.RememberHash = us.hmac.Hash(user.Remember)

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
	hmac := hash.NewHMAC(hmacSecretKey)
	return &UserService{
		db:   db,
		hmac: hmac,
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
	db   *gorm.DB
	hmac hash.HMAC
}
