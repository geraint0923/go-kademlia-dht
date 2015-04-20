package kademlia

import (
	"container/list"
	"fmt"
)

type KBucket struct {
	list.List
}

func NewKBucket() *KBucket {
	ret := &KBucket{}
	ret.Init()
	return ret
}

func (b *KBucket) FindContact(nodeId ID) (*list.Element, error) {
	fmt.Println("find => " + nodeId.AsString())
	for e := b.Front(); e != nil; e = e.Next() {
		c := e.Value.(Contact)
		if nodeId.Equals(c.NodeID) {
			return e, nil
		}
	}
	return nil, &NotFoundError{nodeId, "KBucket not found"}
}

func (b *KBucket) GetLast(kk int) (ret []Contact) {
	ret = []Contact{}
	for e := b.Back(); e != nil && kk > 0; e = e.Prev() {
		ret = append(ret, e.Value.(Contact))
		kk--
	}
	return
}

func (b *KBucket) Full() bool {
	return b.Len() >= K
}
