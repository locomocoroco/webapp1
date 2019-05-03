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

type UserDB interface {
	ByID(id uint) (*Users, error)
	ByEmail(email string) (*Users, error)
	ByRemember(token string) (*Users, error)

	Create(user *Users) error
	Update(user *Users) error
	Delete(id uint) error

	Close() error

	AutoMigrate() error
	DestructiveReset() error
}

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

func (ug *userGorm) ByID(id uint) (*Users, error) {
	var user Users
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

func (ug *userGorm) ByEmail(email string) (*Users, error) {
	var user Users
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err

}
func (ug *userGorm) ByRemember(token string) (*Users, error) {
	var user Users
	rememberHash := ug.hmac.Hash(token)
	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
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
func (ug *userGorm) Update(user *Users) error {
	if user.Remember != "" {
		user.RememberHash = ug.hmac.Hash(user.Remember)
	}
	return ug.db.Save(user).Error
}
func (ug *userGorm) Create(user *Users) error {
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

	user.RememberHash = ug.hmac.Hash(user.Remember)

	return ug.db.Create(user).Error
}
func (ug *userGorm) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := Users{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

//NewUserService
func NewUserService(connectionInfo string) (*UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}
	return &UserService{
		UserDB: &userValidator{
			UserDB: ug,
		},
	}, nil
}

//Closes db conn
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.DropTableIfExists(&Users{}).Error; err != nil {
		return err
	}
	return ug.AutoMigrate()
}
func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&Users{}).Error; err != nil {
		return err
	}
	return nil
}
func (ug *userGorm) Close() error {
	return ug.db.Close()
}

type UserService struct {
	UserDB
}
type userValidator struct {
	UserDB
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	hmac := hash.NewHMAC(hmacSecretKey)
	return &userGorm{
		db:   db,
		hmac: hmac,
	}, nil
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db   *gorm.DB
	hmac hash.HMAC
}
