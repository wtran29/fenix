package webdavfilesystem

import "github.com/wtran29/fenix/cmd/filesystems"

type WebDAV struct {
	Host string
	User string
	Pass string
}

func (s *WebDAV) Put(filename, folder string) error {
	return nil
}

func (s *WebDAV) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing
	return listing, nil
}

func (s *WebDAV) Delete(itemsToDel []string) bool {
	return true
}

func (s *WebDAV) Get(destination string, items ...string) error {
	return nil
}
