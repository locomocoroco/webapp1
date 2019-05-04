package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
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
type UserService interface {
	Auth(email, password string) (*Users, error)
	UserDB
}

var (
	ErrNotFound  = errors.New("resource not found")
	ErrInvalidID = errors.New("invalid id given")
	ErrInvalidPW = errors.New("invalid password provided")
)
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`)

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
func (ug *userGorm) ByRemember(rememberHash string) (*Users, error) {
	var user Users

	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (us *userService) Auth(email, password string) (*Users, error) {
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
	return ug.db.Save(user).Error
}
func (ug *userGorm) Create(user *Users) error {
	return ug.db.Create(user).Error
}
func (ug *userGorm) Delete(id uint) error {
	user := Users{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

//NewUserService
func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := &userValidator{
		hmac:   hmac,
		UserDB: ug,
	}
	return &userService{
		UserDB: &userValidator{
			UserDB: uv,
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

var _ UserService = &userService{}

type userService struct {
	UserDB
}

var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
	hmac hash.HMAC
}

func (uv *userValidator) ByEmail(email string) (*Users, error) {
	user := Users{
		Email: email,
	}
	if err := runUserValFuncs(&user, uv.normalizeEmail); err != nil {
		return nil, err
	}
	return uv.UserDB.ByEmail(user.Email)
}

func (uv *userValidator) ByRemember(token string) (*Users, error) {
	user := Users{
		Remember: token,
	}
	if err := runUserValFuncs(&user, uv.hmacRemember); err != nil {
		return nil, err
	}

	return uv.UserDB.ByRemember(user.RememberHash)
}
func (uv *userValidator) Create(user *Users) error {
	if err := runUserValFuncs(user, uv.bcryptPassword, uv.setRememberIfUnset,
		uv.hmacRemember, uv.normalizeEmail, uv.requireEmail, uv.emailFormat, uv.emailIsAvail); err != nil {
		return err
	}
	return uv.UserDB.Create(user)
}

type userValFunc func(*Users) error

func runUserValFuncs(user *Users, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

func (uv *userValidator) bcryptPassword(user *Users) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password + userPepperPw)
	passwordHash, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(passwordHash)
	user.Password = ""
	return nil
}
func (uv *userValidator) hmacRemember(user *Users) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}
func (uv *userValidator) setRememberIfUnset(user *Users) error {
	if user.Remember != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}
func (uv *userValidator) idGreaterThan(n uint) userValFunc {
	return userValFunc(func(user *Users) error {
		if user.ID <= 0 {
			return ErrInvalidID
		}
		return nil
	})

}
func (uv *userValidator) normalizeEmail(user *Users) error {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	return nil
}
func (uv *userValidator) requireEmail(user *Users) error {
	if user.Email == "" {
		return errors.New("email is required")
	}
	return nil
}
func (uv *userValidator) emailFormat(user *Users) error {
	if user.Email == "" {
		return nil
	}
	if !emailRegex.MatchString(user.Email) {
		return errors.New("email is not valid")
	}
	return nil
}
func (uv *userValidator) emailIsAvail(user *Users) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if user.ID != existing.ID {
		return errors.New("email is already taken")
	}
	return nil
}
func (uv *userValidator) Update(user *Users) error {
	if err := runUserValFuncs(user, uv.bcryptPassword, uv.hmacRemember,
		uv.normalizeEmail, uv.requireEmail, uv.emailFormat, uv.emailIsAvail); err != nil {
		return err
	}
	return uv.UserDB.Update(user)
}
func (uv *userValidator) Delete(id uint) error {
	var user Users
	user.ID = id
	if err := runUserValFuncs(&user, uv.idGreaterThan(0)); err != nil {
		return err
	}
	return uv.UserDB.Delete(id)
}
func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &userGorm{
		db: db,
	}, nil
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db *gorm.DB
}
