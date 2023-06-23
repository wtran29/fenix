package fenix

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/gabriel-vasile/mimetype"
	"github.com/wtran29/fenix/fenix/cmd/filesystems"
)

func (f *Fenix) UploadFile(r *http.Request, destination, field string, fs filesystems.FS) error {
	filename, err := f.getFileToUpload(r, field)
	if err != nil {
		f.ErrorLog.Println(err)
		return err
	}

	if fs != nil {
		err = fs.Put(filename, destination)
		if err != nil {
			f.ErrorLog.Println(err)
			return err
		}
	} else {
		err = os.Rename(filename, fmt.Sprintf("%s/%s", destination, path.Base(filename)))
		if err != nil {
			f.ErrorLog.Println(err)
			return err
		}
	}

	defer func() {
		_ = os.Remove(filename)
	}()

	return nil
}

func (f *Fenix) getFileToUpload(r *http.Request, fieldname string) (string, error) {
	err := r.ParseMultipartForm(f.config.uploads.maxUploadSize)
	if err != nil {
		fmt.Println("could not parse multipart form:", err)
	}
	file, header, err := r.FormFile(fieldname)
	if err != nil {
		return "", err
	}

	defer file.Close()
	mimeType, err := mimetype.DetectReader(file)
	if err != nil {
		return "", err
	}

	// go back to start of file
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", err
	}

	if !inSlice(f.config.uploads.allowedMimeTypes, mimeType.String()) {
		return "", errors.New("invalid file type uploaded")
	}

	dst, err := os.Create(fmt.Sprintf("./tmp/%s", header.Filename))
	if err != nil {
		return "", err
	}

	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("./tmp/%s", header.Filename), nil
}

func inSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false

}
