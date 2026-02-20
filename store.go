package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key)) // [20] byte -> []byte -> [:] (to be slice)

	hashString := hex.EncodeToString(hash[:])

	blockSize := 5

	sliceLength := len(hashString) / blockSize
	paths := make([]string, sliceLength)

	for i := 0; i < sliceLength; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashString[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		Original: hashString,
	}
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	Original string
}

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

func (p PathKey) Filename() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Original)
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	pathAndFilename := pathKey.Filename()

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
