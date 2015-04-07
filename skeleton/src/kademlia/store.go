package kademlia

import (
	"fmt"
)

type StorageError struct {
	Info string
}

type Storage struct {
	Values map[ID][]byte
}

func NewStorage() *Storage {
	res := &Storage{
		Values: make(map[ID][]byte),
	}
	return res
}

func (s *Storage) Retrieve(key ID) (val []byte) {
	val = nil
	fmt.Println("Retrive Key: " + key.AsString())
	if v, ok := s.Values[key]; ok {
		val = v
	}
	return
}

func (s *Storage) Store(key ID, val []byte) (retErr error) {
	retErr = nil
	if _, ok := s.Values[key]; !ok {
		s.Values[key] = val
	} else {
		retErr = StorageError{"Value exists!"}
	}
	return
}

func (se StorageError) Error() string {
	return se.Info
}
