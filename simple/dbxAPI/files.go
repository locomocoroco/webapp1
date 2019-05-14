package dbxAPI

import (
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	dbxf "github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

type Folder struct {
	Name string
	Path string
}

type File struct {
	Name string
	Path string
}

func List(accessToken, path string) ([]Folder, []File, error) {
	config := dropbox.Config{
		Token: accessToken,
	}
	dbx := dbxf.New(config)
	res, err := dbx.ListFolder(&dbxf.ListFolderArg{
		Path: path,
	})
	if err != nil {
		return nil, nil, err
	}
	var folders []Folder
	var files []File

	for _, entry := range res.Entries {
		switch meta := entry.(type) {
		case *dbxf.FolderMetadata:
			folders = append(folders, Folder{
				Name: meta.Name,
				Path: meta.PathLower,
			})
		case *dbxf.FileMetadata:
			files = append(files, File{
				Name: meta.Name,
				Path: meta.PathLower,
			})
		}
	}
	return folders, files, nil
}
