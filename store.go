package main

import (
	"io"
	"log"
	"os"
)

type PathTransformFunc func(string) string
type StoreOptions struct {
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) string {
	return key
}

type Store struct {
	StoreOptions
}

func NewStore(options StoreOptions) *Store {
	return &Store{
		StoreOptions: options,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathName := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}

	fileName := "someFileName"

	pathAndFilename := pathName + "/" + fileName

	f, err := os.Create(pathAndFilename)

	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("wrote %d bytes to disk: %s", n, pathAndFilename)

	return nil
}
