package kademlia

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

// for Kademlia specific arguement
const (
	alpha        = 3
	B            = IDBytes * 8
	K            = 20
	T_EXPIRE     = 86400
	T_REFRESH    = 3600
	T_REPLICATE  = 3600
	T_REPUBULISH = 86400
)

// for general
const (
	OP_OK      = iota
	OP_FAILURE = iota
)

// for RouteTable operation
const (
	RT_LEN                = iota // Value is not required
	RT_BUCKET_LEN         = iota // Value should be the index of buckets, from 0
	RT_BUCKET_HEAD        = iota // Value should be the index of buckets
	RT_GET_K_LAST         = iota // Value should be the request ID
	RT_REMOVE_BUCKET_HEAD = iota // Value should be the removed Contact
	RT_SPLIT_AND_ADD      = iota // Value should be the inserted Contact
	RT_MOVE_HEAD_TO_TAIL  = iota // Value should be the moved Contact
)

type msg struct {
	Opcode int
	Value  interface{}
}

// Core Kademlia type. You can put whatever state you want in this.
type Kademlia struct {
	NodeID                    ID
	LocalStorage              *Storage
	routeTable                []*KBucket
	routeTableChannel         chan *msg
	routeTableResponseChannel chan *msg
	exitChannel               chan bool
}

func NewKademlia() (k *Kademlia) {
	// TODO: Assign yourself a random ID and prepare other state here.
	k = new(Kademlia)
	k.NodeID = NewRandomID()
	k.LocalStorage = NewStorage()
	k.routeTable = []*KBucket{}
	kb := NewKBucket(0)
	k.routeTable = append(k.routeTable, kb)
	k.routeTableChannel = make(chan *msg)
	k.routeTableResponseChannel = make(chan *msg)
	k.exitChannel = make(chan bool)
	fmt.Println("Init Done")
	return
}

func Run(k *Kademlia) {
	go HandleRouteTable(k)
}

func Stop(k *Kademlia) {
	close(k.exitChannel)
}

func HandleRouteTable(k *Kademlia) {
	select {
	case <-k.exitChannel:
		break
	case m := <-k.routeTableChannel:
		fmt.Println(m)
		switch m.Opcode {
		case RT_LEN:
			k.routeTableResponseChannel <- &msg{
				Opcode: OP_OK,
				Value:  len(k.routeTable),
			}
		case RT_BUCKET_LEN:
			idx := m.Value.(int)
			res := &msg{
				Opcode: OP_FAILURE,
				Value:  nil,
			}
			if idx < len(k.routeTable) {
				res.Opcode = OP_OK
				res.Value = k.routeTable[idx].Len()
			}
			k.routeTableResponseChannel <- res
		case RT_BUCKET_HEAD:
			idx := m.Value.(int)
			res := &msg{
				Opcode: OP_FAILURE,
				Value:  nil,
			}
			if idx < len(k.routeTable) {
				res.Opcode = OP_OK
				res.Value = k.routeTable[idx].Head()
			}
			k.routeTableResponseChannel <- res
		case RT_GET_K_LAST:
			id := m.Value.(ID)
			idx := k.NodeID.Xor(id).PrefixLen()
			if idx >= len(k.routeTable) {
				idx = len(k.routeTable) - 1
			}
			res := &msg{
				Opcode: OP_FAILURE,
				Value:  nil,
			}
			if idx >= 0 {
				ret := []*Contact{}
				for left := K; left > 0 && idx >= 0; {
					tmp := k.routeTable[idx].GetKLast(left)
					left -= len(tmp)
					idx--
					ret = append(ret, tmp...)
				}
				res.Opcode = OP_OK
				res.Value = ret
			}
			k.routeTableResponseChannel <- res
		case RT_REMOVE_BUCKET_HEAD:
			c := m.Value.(*Contact)
			idx := k.NodeID.Xor(c.NodeID).PrefixLen()
			if idx >= len(k.routeTable) {
				idx = len(k.routeTable) - 1
			}
			res := &msg{
				Opcode: OP_FAILURE,
				Value:  nil,
			}
			if idx >= 0 && k.routeTable[idx].Len() >= K {
				res.Opcode = OP_OK
				res.Value = k.routeTable[idx].RemoveHead(c)
			}
			k.routeTableResponseChannel <- res
		case RT_SPLIT_AND_ADD:
			c := m.Value.(*Contact)
			res := &msg{
				Opcode: OP_FAILURE,
				Value:  nil,
			}
			idx := k.NodeID.Xor(c.NodeID).PrefixLen()
			split := false
			if idx == B {
				idx = -1
			}
			tmp := idx
			if idx >= len(k.routeTable) {
				idx = len(k.routeTable) - 1
				split = true
			}
			if idx >= 0 && k.routeTable[idx].Len() < K {
				res.Opcode = OP_OK
				res.Value = true
				k.routeTable[idx].Append(c)
			} else if idx >= 0 && split {
				res.Opcode = OP_OK
				newBucket := k.routeTable[idx].Split(k.NodeID)
				k.routeTable = append(k.routeTable, newBucket)
				if tmp > idx {
					if k.routeTable[idx+1].Len() < K {
						k.routeTable[idx+1].Append(c)
						res.Value = true
					}
				} else {
					if k.routeTable[idx].Len() < K {
						k.routeTable[idx].Append(c)
						res.Value = true
					}
				}
			} else {
				res.Opcode = OP_OK
				res.Value = false
			}
			k.routeTableResponseChannel <- res
		case RT_MOVE_HEAD_TO_TAIL:
			c := m.Value.(*Contact)
			res := &msg{
				Opcode: OP_FAILURE,
				Value:  nil,
			}
			idx := k.NodeID.Xor(c.NodeID).PrefixLen()
			if idx == B {
				idx = -1
			}
			if idx >= len(k.routeTable) {
				idx = len(k.routeTable) - 1
			}
			if idx >= 0 {
				res.Opcode = OP_OK
				res.Value = true
				k.routeTable[idx].MoveHeadToTail(c)
			}
			k.routeTableResponseChannel <- res
		default:
			log.Fatal("HandleRouteTable uknown opcode: " + strconv.Itoa(m.Opcode))
		}
		break
	}
}

func DoPing(k *Kademlia, remoteHost net.IP, port uint16) {
}

func DoStore(k *Kademlia, remoteContact *Contact, key ID, value []byte) {
}

func DoFindValue(k *Kademlia, remoteContact *Contact, key ID) {
}

func DoFindNode(k *Kademlia, remoteContact *Contact, searchKey ID) {
}

func IterativeFindNode(k *Kademlia, key ID) {
}

func IterativeFindValue(k *Kademlia, key ID) {
}

func IterativeStore(k *Kademlia, key ID, value []byte) {
}
