package models

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"
)

const OAtuhDropbox = "dropbox"

type OAuth struct {
	gorm.Model
	UserID  uint   `gorm:"not null;unique_index:services"`
	Service string `gorm:"not null;unique_index:services"`
	oauth2.Token
}

func NewOAuthService(db *gorm.DB) OAuthService {
	return &oauthValidator{OAuthDB: &oauthGorm{db: db}}
}

type OAuthDB interface {
	Find(userID uint, service string) (*OAuth, error)
	Create(oauth *OAuth) error
	Delete(userID uint) error
}
type OAuthService interface {
	OAuthDB
}

var _ OAuthDB = &oauthGorm{}

type oauthValidator struct {
	OAuthDB
}

type oauthGorm struct {
	db *gorm.DB
}

func (og *oauthGorm) Find(userID uint, service string) (*OAuth, error) {
	var oauth OAuth
	err := og.db.Where("user_id=? AND service=?", userID, service).First(&oauth).Error
	if err != nil {
		return nil, err
	}
	return &oauth, nil
}
func (og *oauthGorm) Create(oauth *OAuth) error {
	return og.db.Create(oauth).Error
}
func (og *oauthGorm) Delete(id uint) error {
	oauth := OAuth{Model: gorm.Model{ID: id}}
	return og.db.Unscoped().Delete(&oauth).Error
}
func (ov *oauthValidator) Create(oauth *OAuth) error {
	if err := runoauthValFuncs(oauth, ov.userIDxServRequired, ov.tokenRequired); err != nil {
		return err
	}
	return ov.OAuthDB.Create(oauth)
}
func (ov *oauthValidator) Delete(userID uint) error {
	if userID <= 0 {
		return ErrIDInvalid
	}
	return ov.OAuthDB.Delete(userID)
}
func (ov *oauthValidator) userIDxServRequired(o *OAuth) error {
	if o.UserID <= 0 {
		return ErrUserIDRequired
	}
	if o.Service == "" {
		return ErrNotFound
	}
	return nil
}
func (ov *oauthValidator) tokenRequired(o *OAuth) error {
	if o.Token.AccessToken == "" {
		return ErrNotFound
	}
	return nil
}

type oauthValFunc func(*OAuth) error

func runoauthValFuncs(oauth *OAuth, fns ...oauthValFunc) error {
	for _, fn := range fns {
		if err := fn(oauth); err != nil {
			return err
		}
	}
	return nil
}
