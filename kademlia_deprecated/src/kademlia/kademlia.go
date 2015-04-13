package kademlia

import (
	"fmt"
	//"log"
	"net"
	"net/rpc"
	"strconv"
	"sync"
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
/*
const (
	RT_LEN          = iota // Value is not required
	RT_BUCKET_LEN   = iota // Value should be the index of buckets, from 0
	RT_BUCKET_HEAD  = iota // Value should be the index of buckets
	RT_GET_K_LAST   = iota // Value should be the request ID
	RT_REMOVE       = iota // Value should be the removed Contact
	RT_SPLIT_ADD    = iota // Value should be the inserted Contact
	RT_MOVE_TO_TAIL = iota // Value should be the moved Contact
)

type msg struct {
	Opcode int
	Value  interface{}
}
*/

// Core Kademlia type. You can put whatever state you want in this.
type Kademlia struct {
	NodeID          ID
	LocalStorage    *Storage
	routeTable      []*KBucket
	routeTableMutex *sync.Mutex
	exitChannel     chan bool
	//routeTableChannel         chan *msg
	//routeTableResponseChannel chan *msg
}

func NewKademlia() (k *Kademlia) {
	// TODO: Assign yourself a random ID and prepare other state here.
	k = new(Kademlia)
	k.NodeID = NewRandomID()
	k.LocalStorage = NewStorage()
	k.routeTable = []*KBucket{}
	k.routeTableMutex = &sync.Mutex{}
	kb := NewKBucket(0)
	k.routeTable = append(k.routeTable, kb)
	//k.routeTableChannel = make(chan *msg)
	//k.routeTableResponseChannel = make(chan *msg)
	k.exitChannel = make(chan bool)
	fmt.Println("Init Done")
	return
}

func Run(k *Kademlia) {
	//go HandleRouteTable(k)
}

func Stop(k *Kademlia) {
	close(k.exitChannel)
}

/*
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
		case RT_REMOVE:
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
				res.Value = k.routeTable[idx].Remove(c)
			}
			k.routeTableResponseChannel <- res
		case RT_SPLIT_ADD:
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
		case RT_MOVE_TO_TAIL:
			c := m.Value.(*Contact)
			res := &msg{
				Opcode: OP_FAILURE,
				Value:  false,
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
				res.Value = k.routeTable[idx].MoveToTail(c)
			}
			k.routeTableResponseChannel <- res
		default:
			log.Fatal("HandleRouteTable uknown opcode: " + strconv.Itoa(m.Opcode))
		}
		break
	}
}
*/

// update the route table using given contact
func UpdateRouteTable(k *Kademlia, c *Contact) {
	// TODO: use internalDoPing to test survival of the Contact
	idx := k.NodeID.Xor(c.NodeID).PrefixLen()

	k.routeTableMutex.Lock()
	curIdx := idx
	var head *Contact
	var headIdx int
	var success bool
	if curIdx >= len(k.routeTable) {
		curIdx = len(k.routeTable) - 1
	}
	if k.routeTable[curIdx].MoveToTail(c) {
		goto done
	}
	if k.routeTable[curIdx].Len() < K {
		k.routeTable[curIdx].Append(c)
		goto done
	}
	for idx >= len(k.routeTable) {
		newBucket := k.routeTable[curIdx].Split(k.NodeID)
		k.routeTable = append(k.routeTable, newBucket)
		curIdx++
		if k.routeTable[curIdx].Len() < K {
			k.routeTable[curIdx].Append(c)
			goto done
		}
	}
	head = k.routeTable[curIdx].Head()

	k.routeTableMutex.Unlock()
	success = internalDoPing(k, head.Host, head.Port, false)
	k.routeTableMutex.Lock()

	headIdx = k.NodeID.Xor(head.NodeID).PrefixLen()
	if headIdx >= len(k.routeTable) {
		headIdx = len(k.routeTable) - 1
	}
	if success {
		k.routeTable[headIdx].MoveToTail(head)
	} else {
		k.routeTable[headIdx].Remove(head)
	}
	if k.routeTable[idx].Len() < K {
		k.routeTable[idx].Append(c)
	}
done:
	k.routeTableMutex.Unlock()
}

func getRPCClient(remoteHost net.IP, port uint16) *rpc.Client {
	cli, err := rpc.DialHTTP("tcp", remoteHost.String()+":"+strconv.Itoa(int(port)))
	if err != nil {
		return nil
	}
	return cli
}

func internalDoPing(k *Kademlia, remoteHost net.IP, port uint16, update bool) bool {
	// TODO: begin to do ping to other server
	client := getRPCClient(remoteHost, port)
	if client == nil {
		fmt.Println("Nil client")
		return false
	}
	ping := new(Ping)
	ping.MsgID = NewRandomID()
	var pong Pong
	err := client.Call("Kademlia.Ping", ping, &pong)
	if err != nil {
		fmt.Println("Ping failed")
		return false
	}
	if update {
		go UpdateRouteTable(k, &Contact{
			Host: remoteHost,
			Port: port,
		})
	}
	return true
}

func DoPing(k *Kademlia, remoteHost net.IP, port uint16) bool {
	success := internalDoPing(k, remoteHost, port, true)
	if success {
		return true
	} else {
		return false
	}
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
