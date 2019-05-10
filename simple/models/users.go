package models

import (
	"github.com/jinzhu/gorm"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
	"webapp1/simple/hash"
	"webapp1/simple/rand"
)

type UserDB interface {
	ByID(id uint) (*Users, error)
	ByEmail(email string) (*Users, error)
	ByRemember(token string) (*Users, error)

	Create(user *Users) error
	Update(user *Users) error
	Delete(id uint) error
}
type UserService interface {
	Auth(email, password string) (*Users, error)
	InitiateReset(userID uint) (string, error)
	CompleteReset(token, newPw string) (*Users, error)
	UserDB
}

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
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+us.pepper))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrPasswordIncorrect
		default:
			return nil, err
		}
	}
	return foundUser, nil
}
func (us *userService) InitiateReset(userID uint) (string, error) {
	pwr := pwReset{UserID: userID}
	if err := us.pwResetDB.Create(&pwr); err != nil {
		return "", nil
	}
	return pwr.Token, nil
}
func (us *userService) CompleteReset(token, newPw string) (*Users, error) {
	pwr, err := us.pwResetDB.ByToken(token)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrTokenInvalid
		}
	}
	if time.Now().Sub(pwr.CreatedAt) > (6 * time.Hour) {
		return nil, ErrTokenInvalid
	}
	user, err := us.ByID(pwr.ID)
	if err != nil {
		return nil, err
	}
	err = us.Update(user)
	if err != nil {
		return nil, err
	}
	err = us.pwResetDB.Delete(pwr.ID)
	if err != nil {
		log.Println(err)
	}
	return user, nil
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
func newUserService(db *gorm.DB, hmacSecretKey, pepper string) UserService {
	ug := &userGorm{db}
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := &userValidator{
		hmac:   hmac,
		UserDB: ug,
		pepper: pepper,
	}
	return &userService{
		UserDB: &userValidator{
			UserDB: uv,
		},
		pepper:    pepper,
		pwResetDB: newPwResetValidator(&pwResetGorm{db}, hmac),
	}
}

var _ UserService = &userService{}

type userService struct {
	UserDB
	pepper    string
	pwResetDB pwResetDB
}

var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
	hmac   hash.HMAC
	pepper string
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
	if err := runUserValFuncs(user, uv.requirePassword,
		uv.lengthPassword,
		uv.bcryptPassword,
		uv.requirePasswordHash,
		uv.setRememberIfUnset,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.requireRememberHash,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail); err != nil {
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
	pwBytes := []byte(user.Password + uv.pepper)
	passwordHash, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(passwordHash)
	user.Password = ""
	return nil
}
func (uv *userValidator) lengthPassword(user *Users) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}
func (uv *userValidator) requirePassword(user *Users) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}
func (uv *userValidator) requirePasswordHash(user *Users) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}
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
func (uv *userValidator) rememberMinBytes(user *Users) error {
	if user.Remember == "" {
		return nil
	}
	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return ErrRememberTooShort
	}
	return nil
}
func (uv *userValidator) requireRememberHash(user *Users) error {
	if user.RememberHash == "" {
		return ErrRememberRequired
	}
	return nil
}
func (uv *userValidator) idGreaterThan(n uint) userValFunc {
	return userValFunc(func(user *Users) error {
		if user.ID <= 0 {
			return ErrIDInvalid
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
		return ErrEmailRequired
	}
	return nil
}
func (uv *userValidator) emailFormat(user *Users) error {
	if user.Email == "" {
		return nil
	}
	if !emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
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
		return ErrEmailTaken
	}
	return nil
}
func (uv *userValidator) Update(user *Users) error {
	if err := runUserValFuncs(user, uv.lengthPassword,
		uv.bcryptPassword,
		uv.requirePasswordHash,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.requireRememberHash,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail); err != nil {
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

var _ UserDB = &userGorm{}

type userGorm struct {
	db *gorm.DB
}
