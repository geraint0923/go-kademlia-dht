package kademlia

import (
	"container/list"
)

type KBucket struct {
	PrefixLen   int
	ContactList *list.List
}

func NewKBucket(prefixLen int) (b *KBucket) {
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
	if b.PrefixLen >= IDBytes*8 {
		return
	}
	ret = NewKBucket(b.PrefixLen + 1)
	oldList := b.ContactList
	b.ContactList = list.New()
	for e := oldList.Front(); e != nil; e = e.Next() {
		contact := e.Value.(*Contact)
		if contact.NodeID.TestBit(nodeID, b.PrefixLen) {
			ret.ContactList.PushBack(contact)
		} else {
			b.ContactList.PushBack(contact)
		}
	}
	return
}

func (b *KBucket) Head() (ret *Contact) {
	ret = nil
	head := b.ContactList.Front()
	if head != nil {
		ret = head.Value.(*Contact)
	}
	return
}

func (b *KBucket) Remove(c *Contact) bool {
	/*
		head := b.ContactList.Front()
		if head != nil && c.Equals(head.Value.(*Contact)) {
			b.ContactList.Remove(head)
			return true
		}
		return false
	*/
	for e := b.ContactList.Front(); e != nil; e = e.Next() {
		if c.Equals(e.Value.(*Contact)) {
			b.ContactList.Remove(e)
			return true
		}
	}
	return false
}

func (b *KBucket) MoveToTail(c *Contact) bool {
	/*
		head := b.ContactList.Front()
		if head != nil && c.Equals(head.Value.(*Contact)) {
			b.ContactList.MoveToBack(head)
		}
	*/
	for e := b.ContactList.Front(); e != nil; e = e.Next() {
		if c.Equals(e.Value.(*Contact)) {
			b.ContactList.MoveToBack(e)
			return true
		}
	}
	return false
}

func (b *KBucket) Append(c *Contact) {
	b.ContactList.PushBack(c)
}

func (b *KBucket) GetKLast(k int) (ret []*Contact) {
	ret = []*Contact{}
	c := 0
	for e := b.ContactList.Back(); e != nil && c < k; e = e.Prev() {
		c++
		ret = append(ret, e.Value.(*Contact))
	}
	return
}
