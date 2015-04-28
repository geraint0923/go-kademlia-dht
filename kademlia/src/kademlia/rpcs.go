package kademlia

// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.

import (
	"errors"
	//	"fmt"
	"net"
	//	"strconv"
)

type KademliaCore struct {
	kademlia *Kademlia
}

// Host identification.
type Contact struct {
	NodeID ID
	Host   net.IP
	Port   uint16
}

///////////////////////////////////////////////////////////////////////////////
// PING
///////////////////////////////////////////////////////////////////////////////
type PingMessage struct {
	Sender Contact
	MsgID  ID
}

type PongMessage struct {
	MsgID  ID
	Sender Contact
}

func (kc *KademliaCore) Ping(ping PingMessage, pong *PongMessage) error {
	// TODO: Finish implementation
	pong.MsgID = CopyID(ping.MsgID)
	// Specify the sender
	pong.Sender = kc.kademlia.SelfContact
	// Update contact, etc
	kc.kademlia.updateChannel <- ping.Sender
	//fmt.Println("hehe: " + ping.Sender.Host.String() + ":" + strconv.Itoa(int(ping.Sender.Port)))
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STORE
///////////////////////////////////////////////////////////////////////////////
type StoreRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
	Value  []byte
}

type StoreResult struct {
	MsgID ID
	Err   error
}

func (kc *KademliaCore) Store(req StoreRequest, res *StoreResult) error {
	// TODO: Implement.
	res.MsgID = req.MsgID
	ok := kc.kademlia.storage.Put(req.Key, req.Value)
	if ok {
		res.Err = nil
	} else {
		res.Err = errors.New("Failed to store")
	}
	kc.kademlia.updateChannel <- req.Sender
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_NODE
///////////////////////////////////////////////////////////////////////////////
type FindNodeRequest struct {
	Sender Contact
	MsgID  ID
	NodeID ID
}

type FindNodeResult struct {
	MsgID ID
	Nodes []Contact
	Err   error
}

func (kc *KademliaCore) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	// TODO: Implement.
	res.MsgID = req.MsgID
	res.Nodes = filterContactList(kc.kademlia.getLastContactFromRoutingTable(req.NodeID), req.Sender.NodeID)
	kc.kademlia.updateChannel <- req.Sender
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_VALUE
///////////////////////////////////////////////////////////////////////////////
type FindValueRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
}

// If Value is nil, it should be ignored, and Nodes means the same as in a
// FindNodeResult.
type FindValueResult struct {
	MsgID ID
	Value []byte
	Nodes []Contact
	Err   error
}

func (kc *KademliaCore) FindValue(req FindValueRequest, res *FindValueResult) error {
	// TODO: Implement.
	res.MsgID = req.MsgID
	val, ok := kc.kademlia.storage.Get(req.Key)
	if ok {
		res.Value = val
		res.Nodes = nil
	} else {
		res.Value = nil
		res.Nodes = filterContactList(kc.kademlia.getLastContactFromRoutingTable(req.Key), req.Sender.NodeID)
	}
	res.Err = nil
	kc.kademlia.updateChannel <- req.Sender
	return nil
}
