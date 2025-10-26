package storage

import (
	"os"
	"syscall"
)

type MappedFile struct {
	Data []byte
	File *os.File
}

func Mmap(path string) (*MappedFile, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(stat.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		f.Close()
		return nil, err
	}

	return &MappedFile{
		Data: data,
		File: f,
	}, nil
}

func (m *MappedFile) Close() error {
	if err := syscall.Munmap(m.Data); err != nil {
		m.File.Close()
		return err
	}
	syscall.Munmap(m.Data)
	return m.File.Close()
}
