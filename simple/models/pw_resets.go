package models

import (
	"github.com/jinzhu/gorm"
	"webapp1/simple/hash"
	"webapp1/simple/rand"
)

type pwReset struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	Token     string `gorm:"-"`
	TokenHash string `gorm:"not null;unique_index"`
}
type pwResetDB interface {
	ByToken(t string) (*pwReset, error)
	Create(pwr *pwReset) error
	Delete(id uint) error
}
type pwResetValidator struct {
	pwResetDB
	hmac hash.HMAC
}

func newPwResetValidator(db pwResetDB, hmac hash.HMAC) *pwResetValidator {
	return &pwResetValidator{
		pwResetDB: db,
		hmac:      hmac,
	}
}

type pwResetValFn func(*pwReset) error

func runPwResetValFns(pwr *pwReset, fns ...pwResetValFn) error {
	for _, fn := range fns {
		if err := fn(pwr); err != nil {
			return err
		}
	}
	return nil
}
func (prv *pwResetValidator) ByToken(token string) (*pwReset, error) {
	pwr := pwReset{Token: token}
	err := runPwResetValFns(&pwr, prv.hmacToken)
	if err != nil {
		return nil, err
	}
	return &pwr, err
}
func (prv *pwResetValidator) Create(pwr *pwReset) error {
	err := runPwResetValFns(pwr,
		prv.requireUserID,
		prv.setTokenIfUnset,
		prv.hmacToken,
	)
	if err != nil {
		return err
	}
	return prv.pwResetDB.Create(pwr)
}
func (prv *pwResetValidator) Delete(id uint) error {
	if id <= 0 { //err := runPwResetValFns(pwr,prv.requireUserID,
		return ErrIDInvalid
	}
	return prv.pwResetDB.Delete(id)
}
func (prv *pwResetValidator) requireUserID(pwr *pwReset) error {
	if pwr.UserID <= 0 {
		return ErrIDInvalid
	}
	return nil
}
func (prv *pwResetValidator) setTokenIfUnset(pwr *pwReset) error {
	if pwr.Token != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	pwr.Token = token
	return nil
}
func (prv *pwResetValidator) hmacToken(pwr *pwReset) error {
	if pwr.Token == "" {
		return nil
	}
	pwr.TokenHash = prv.hmac.Hash(pwr.Token)
	return nil
}

type pwResetGorm struct {
	db *gorm.DB
}

func (prg *pwResetGorm) ByToken(thash string) (*pwReset, error) {
	var pwr pwReset
	err := first(prg.db.Where("token_hash = ?", thash), &pwr)
	if err != nil {
		return nil, err
	}
	return &pwr, nil
}
func (prg *pwResetGorm) Create(pwr *pwReset) error {
	return prg.db.Create(pwr).Error
}
func (prg *pwResetGorm) Delete(id uint) error {
	pwr := pwReset{
		Model: gorm.Model{ID: id},
	}
	return prg.db.Delete(&pwr).Error
}
