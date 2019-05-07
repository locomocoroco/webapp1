package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ImageService interface {
	Create(galleryID uint, r io.ReadCloser, filename string) error
	ByGalleryID(galleryID uint) ([]Image, error)
	Delete(i *Image) error
}

func NewImageService() ImageService {
	return &imageService{}
}

type Image struct {
	GalleryID uint
	Filename  string
}

func (i *Image) Path() string {
	return "/" + i.RelPath()
}
func (i *Image) RelPath() string {
	return fmt.Sprintf("images/galleries/%v/%v", i.GalleryID, i.Filename)
}

type imageService struct {
}

func (is *imageService) Create(galleryID uint, r io.ReadCloser, filename string) error {
	defer r.Close()
	path, err := is.imagePath(galleryID)
	if err != nil {
		return err
	}
	dst, err := os.Create(path + filename)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}
func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	path := is.path(galleryID)
	stringf, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}
	ret := make([]Image, len(stringf))
	for i := range stringf {
		stringf[i] = strings.Replace(stringf[i], path, "", 1)
		ret[i] = Image{
			Filename:  stringf[i],
			GalleryID: galleryID,
		}
	}
	return ret, nil
}
func (is *imageService) Delete(i *Image) error {
	return os.Remove(i.RelPath())
}
func (is *imageService) path(galleryID uint) string {
	return fmt.Sprintf("images/galleries/%v/", galleryID)
}
func (is *imageService) imagePath(galleryID uint) (string, error) {
	galleryPath := is.path(galleryID)
	err := os.MkdirAll(galleryPath, 0755)
	if err != nil {
		return "", err
	}
	return galleryPath, nil
}
