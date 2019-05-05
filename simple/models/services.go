package models

import "github.com/jinzhu/gorm"

func NewServices(connectionInfo string) (*Services, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &Services{
		User: NewUserService(db),
	}, nil
}

type Services struct {
	Gallery GalleryService
	User    UserService
	db      *gorm.DB
}

//Closes db conn
func (s *Services) DestructiveReset() error {
	if err := s.db.DropTableIfExists(&Users{}, &Gallery{}).Error; err != nil {
		return err
	}
	return s.AutoMigrate()
}
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&Users{}, &Gallery{}).Error

}
func (s *Services) Close() error {
	return s.db.Close()
}
