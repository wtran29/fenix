package sftpfilesystem

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/pkg/sftp"
	"github.com/wtran29/fenix/fenix/cmd/filesystems"
	"golang.org/x/crypto/ssh"
)

type SFTP struct {
	Host string
	User string
	Pass string
	Port string
}

func (s *SFTP) getCredentials() (*sftp.Client, error) {
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}
	cwd, err := client.Getwd()
	if err != nil {
		return nil, err
	}
	log.Println("Current working directory:", cwd)

	return client, nil
}

func (s *SFTP) Put(filename, folder string) error {
	client, err := s.getCredentials()
	if err != nil {
		return err
	}
	defer client.Close()

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f2, err := client.Create(fmt.Sprintf("%s/%s", folder, path.Base(filename)))
	if err != nil {
		return err
	}
	defer f2.Close()

	if _, err := io.Copy(f2, f); err != nil {
		return err
	}
	return nil
}

func (s *SFTP) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing
	client, err := s.getCredentials()
	if err != nil {
		return listing, err
	}
	defer client.Close()

	files, err := client.ReadDir(prefix)
	if err != nil {
		return listing, err
	}

	for _, file := range files {
		var item filesystems.Listing

		if !strings.HasPrefix(file.Name(), ".") {
			b := float64(file.Size())
			kb := b / 1024
			mb := kb / 1024
			item.Key = file.Name()
			item.Size = mb
			item.LastModified = file.ModTime()
			item.IsDir = file.IsDir()
			listing = append(listing, item)
		}

	}
	return listing, nil
}

func (s *SFTP) Delete(itemsToDel []string) bool {
	client, err := s.getCredentials()
	if err != nil {
		return false
	}
	defer client.Close()

	for _, item := range itemsToDel {
		err := client.Remove(item)
		if err != nil {
			return false
		}
	}

	return true
}

func (s *SFTP) Get(destination string, items ...string) error {
	client, err := s.getCredentials()
	if err != nil {
		return err
	}

	defer client.Close()

	for _, item := range items {
		err := func() error {
			// create destination file
			dstFile, err := os.Create(fmt.Sprintf("%s/%s", destination, path.Base(item)))
			if err != nil {
				return err
			}
			defer dstFile.Close()

			// open source file
			srcFile, err := client.Open(item)
			if err != nil {
				return err
			}
			defer srcFile.Close()
			// copy src to dst
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return err
			}

			// flush the in-memory copy
			err = dstFile.Sync()
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
