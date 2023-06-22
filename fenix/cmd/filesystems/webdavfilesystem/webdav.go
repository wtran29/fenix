package webdavfilesystem

import (
	"fmt"
	"os"
	"path"

	"github.com/studio-b12/gowebdav"
	"github.com/wtran29/fenix/fenix/cmd/filesystems"
)

type WebDAV struct {
	Host string
	User string
	Pass string
}

func (w *WebDAV) getCredentials() *gowebdav.Client {
	client := gowebdav.NewClient(w.Host, w.User, w.Pass)
	return client
}

func (w *WebDAV) Put(filename, folder string) error {
	client := w.getCredentials()

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = client.WriteStream(fmt.Sprintf("%s/%s", folder, path.Base(filename)), file, 0664)
	if err != nil {
		return err
	}
	return nil
}

func (w *WebDAV) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing
	return listing, nil
}

func (w *WebDAV) Delete(itemsToDel []string) bool {
	return true
}

func (w *WebDAV) Get(destination string, items ...string) error {
	return nil
}
