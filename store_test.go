package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathKey := CASPathTransformFunc(key)
	fmt.Println(key, pathKey)
	expectedOriginalKey := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"

	if pathKey.PathName != expectedPathName {
		t.Errorf("pathKey.PathName is %s, but expected %s", pathKey.PathName, expectedPathName)
	}

	if pathKey.FileName != expectedOriginalKey {
		t.Errorf("pathKey.FileName is %s, but expected %s", pathKey.FileName, expectedOriginalKey)
	}
}

func TestStore(t *testing.T) {
	options := StoreOptions{
		PathTransformFunc: CASPathTransformFunc,
	}
	store := NewStore(options)
	key := "momsbestpicture"
	data := []byte("some jpg bytes")
	if err := store.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	r, err := store.readStream(key)
	if err != nil {
		t.Error(err)
	}

	b, err := ioutil.ReadAll(r)
	fmt.Println(string(b))
	if string(b) != string(data) {
		t.Errorf("readStream returns %s, but expected %s", b, data)
	}
}
