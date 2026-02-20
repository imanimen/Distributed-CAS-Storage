package main

import (
	"bytes"
	"fmt"
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

	if pathKey.Original != expectedOriginalKey {
		t.Errorf("pathKey.Original is %s, but expected %s", pathKey.Original, expectedOriginalKey)
	}
}

func TestStore(t *testing.T) {
	options := StoreOptions{
		PathTransformFunc: CASPathTransformFunc,
	}

	store := NewStore(options)

	data := bytes.NewReader([]byte("some jpg bytes"))
	if err := store.writeStream("mySpecialPicture", data); err != nil {
		t.Error(err)
	}

}
