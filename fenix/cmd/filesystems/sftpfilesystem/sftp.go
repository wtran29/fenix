package sftpfilesystem

import "github.com/wtran29/fenix/cmd/filesystems"

type SFTP struct {
	Host string
	User string
}

func (s *SFTP) Put(filename, folder string) error {
	return nil
}

func (s *SFTP) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing
	return listing, nil
}

func (s *SFTP) Delete(itemsToDel []string) bool {
	return true
}

func (s *SFTP) Get(destination string, items ...string) error {
	return nil
}
