package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type ServicesConfig func(*Services) error

func WithGorm(dialect, connectionInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(dialect, connectionInfo)
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}
func WithUser(pepper, hmacKey string) ServicesConfig {
	return func(s *Services) error {
		s.User = newUserService(s.db, pepper, hmacKey)
		return nil
	}
}
func WithGallery() ServicesConfig {
	return func(s *Services) error {
		s.Gallery = NewGalleryService(s.db)
		return nil
	}
}
func WithImage() ServicesConfig {
	return func(s *Services) error {
		s.Image = NewImageService()
		return nil
	}
}
func WithLog(mode bool) ServicesConfig {
	return func(s *Services) error {
		s.db.LogMode(mode)
		return nil
	}
}
func WithOAuth() ServicesConfig {
	return func(s *Services) error {
		s.OAuth = NewOAuthService(s.db)
		return nil
	}
}
func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, cfg := range cfgs {
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

type Services struct {
	Gallery GalleryService
	User    UserService
	Image   ImageService
	OAuth   OAuthService
	db      *gorm.DB
}

//Closes db conn
func (s *Services) DestructiveReset() error {
	if err := s.db.DropTableIfExists(&Users{}, &Gallery{}, &pwReset{}, OAuth{}).Error; err != nil {
		return err
	}
	return s.AutoMigrate()
}
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&Users{}, &Gallery{}, &pwReset{}, OAuth{}).Error

}
func (s *Services) Close() error {
	return s.db.Close()
}
