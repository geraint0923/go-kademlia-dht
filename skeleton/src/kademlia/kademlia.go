package kademlia

import (
	"fmt"
	"log"
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
	RT_LEN         = iota
	RT_CLEAR       = iota
	RT_SPLIT       = iota
	RT_HEAD        = iota
	RT_GET_K_LAST  = iota
	RT_ADD_CONTACT = iota
	RT_ADD_BUCKET  = iota
)

type msg struct {
	Opcode int
	Value  interface{}
}

// Core Kademlia type. You can put whatever state you want in this.
type Kademlia struct {
	NodeID                    ID
	routeTable                []*KBucket
	routeTableChannel         chan *msg
	routeTableResponseChannel chan *msg
	exitChannel               chan bool
}

func NewKademlia() (k *Kademlia) {
	// TODO: Assign yourself a random ID and prepare other state here.
	k = new(Kademlia)
	k.NodeID = NewRandomID()
	k.routeTable = []*KBucket{}
	kb := NewKBucket(0)
	k.routeTable = append(k.routeTable, kb)
	k.routeTableChannel = make(chan *msg)
	k.routeTableResponseChannel = make(chan *msg)
	fmt.Println("Init Done")
	return
}

func (k *Kademlia) Run() {
	go k.HandleRouteTable()
}

func (k *Kademlia) Stop() {
	close(k.exitChannel)
}

func (k *Kademlia) HandleRouteTable() {
	select {
	case <-k.exitChannel:
		break
	case m := <-k.routeTableChannel:
		fmt.Println(m)
		switch m.Opcode {
		case RT_LEN:
			k.routeTableResponseChannel <- &msg{
				Opcode: OP_OK,
				Value:  nil,
			}
		case RT_HEAD:
			k.routeTableResponseChannel <- &msg{}
		case RT_CLEAR:
			k.routeTableResponseChannel <- &msg{}
		case RT_SPLIT:
			k.routeTableResponseChannel <- &msg{}
		default:
			log.Fatal("HandleRouteTable uknown opcode: " + strconv.Itoa(m.Opcode))
		}
		break
	}
}
