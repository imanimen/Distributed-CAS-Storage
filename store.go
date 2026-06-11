package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// DefaultRootFolderName is the default folder name for the CAS storage root directory.
const DefaultRootFolderName = "casnetwork"

// CASPathTransformFunc transforms a key into a PathKey by hashing the key
// and splitting the hash into path segments of 5 characters each.
// This creates a hierarchical directory structure for storing files.
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
		FileName: hashString,
	}
}

// PathTransformFunc is a function type that transforms a string key into a PathKey.
// It is used to determine the file path for storing data in the CAS system.
type PathTransformFunc func(string) PathKey

// PathKey represents the result of transforming a key into a file path.
// PathName is the hierarchical directory path (e.g., "68044/29f74/181a6")
// FileName is the full filename (e.g., the full hash string).
type PathKey struct {
	PathName string
	FileName string
}

type StoreOptions struct {
	// Root is the folder name of the root, containing all the files/folders of the system
	Root              string
	PathTransformFunc PathTransformFunc
}

// DefaultPathTransformFunc is the default path transformation function.
// It returns the key as both the path name and file name without any transformation.
var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}

// Store is the content-addressable storage system that handles reading,
// writing, and deleting files based on keys.
type Store struct {
	StoreOptions
}

// NewStore creates a new Store with the given options.
// If PathTransformFunc is not provided, it defaults to DefaultPathTransformFunc.
// If Root is not provided, it defaults to DefaultRootFolderName.
func NewStore(options StoreOptions) *Store {
	if options.PathTransformFunc == nil {
		options.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(options.Root) == 0 {
		options.Root = DefaultRootFolderName
	}
	return &Store{
		StoreOptions: options,
	}
}

// FullPath returns the full path by joining PathName and FileName with a "/".
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.FileName)
}

// Read retrieves data from the store by key.
// It returns an io.Reader that can be used to read the data.
// The caller must close the reader after use to release resources.
func (s *Store) Read(key string) (io.Reader, error) {

	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

// readStream returns an io.ReadCloser for the file associated with the key.
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	return os.Open(fullPathWithRoot)
}

// writeStream writes data from an io.Reader to the store using the given key.
// It creates the necessary directory structure and file if they don't exist.
func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return err
	}

	fullPath := pathKey.FullPath()
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, fullPath)

	f, err := os.Create(fullPathWithRoot)

	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("wrote %d bytes to disk: %s", n, fullPathWithRoot)

	return nil
}

// Delete removes the file and directory associated with the key from the store.
func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)
	defer func() {
		log.Printf("deleting [%s]", pathKey.FileName)
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FirstPathName())

	return os.RemoveAll(firstPathNameWithRoot)

	//if err := os.RemoveAll(pathKey.FullPath()); err != nil {
	//	return err
	//}
	return os.RemoveAll(pathKey.FirstPathName())
}

// Exists checks whether a file exists in the store for the given key.
// It returns true if the file exists, false otherwise.
func (s *Store) Exists(key string) bool {
	pathKey := s.PathTransformFunc(key)
	_, err := os.Stat(pathKey.FullPath())
	if errors.Is(err, os.ErrNotExist) {
		return false

	}

	return true

}

// FirstPathName returns the first directory name in the hierarchical path.
// For example, for path "68044/29f74/181a6", it returns "68044".
func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		panic("empty path name")
	}
	return paths[0]
}
