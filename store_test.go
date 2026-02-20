package main

import (
	"bytes"
	"testing"
)

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
