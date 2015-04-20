package kademlia

import (
	"sync"
)

type LocalStorage struct {
	storage map[string][]byte
	lock    sync.Mutex
}

func NewLocalStorage() *LocalStorage {
	res := &LocalStorage{make(map[string][]byte), sync.Mutex{}}
	return res
}

func (ls *LocalStorage) Get(key ID) (res []byte, ok bool) {
	ls.lock.Lock()
	res, ok = ls.storage[key.AsString()]
	ls.lock.Unlock()
	return
}

func (ls *LocalStorage) Put(key ID, val []byte) (ok bool) {
	ok = true
	ls.lock.Lock()
	ls.storage[key.AsString()] = val
	ls.lock.Unlock()
	return
}
