package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathName := CASPathTransformFunc(key)
	fmt.Println(pathName)
	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathName != expectedPathName {
		t.Errorf("PathTransformFunc returned %s instead of %s", pathName, expectedPathName)
	}
}

func TestStore(t *testing.T) {
	options := StoreOptions{
		PathTransformFunc: DefaultPathTransformFunc,
	}

	store := NewStore(options)

	data := bytes.NewReader([]byte("some jpg bytes"))
	if err := store.writeStream("mySpecialPicture", data); err != nil {
		t.Error(err)
	}

}
