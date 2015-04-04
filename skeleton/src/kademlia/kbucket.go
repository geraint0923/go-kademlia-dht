package kademlia

import (
	"container/list"
)

const (
	K = 20
)

type KBucket struct {
	PrefixLen   int
	ContactList *list.List
}

func New(prefixLen int) (b *KBucket) {
	b = new(KBucket)
	b.PrefixLen = prefixLen
	b.ContactList = list.New()
	return
}

func (b *KBucket) Clear() {
	b.ContactList.Init()
}

func (b *KBucket) Len() int {
	return b.ContactList.Len()
}

func (b *KBucket) Split(nodeID ID) (ret *KBucket) {
	ret = nil
	if b.PrefixLen > IDBytes*8 {
		return
	}
	ret = New(b.PrefixLen + 1)
	oldList := b.ContactList
	b.ContactList = list.New()
	for e := oldList.Front(); e != nil; e = e.Next() {
		contact := e.Value.(Contact)
		if contact.NodeID.TestBit(nodeID, b.PrefixLen) {
			ret.ContactList.PushBack(contact)
		} else {
			b.ContactList.PushBack(contact)
		}
	}
	return
}
