package kademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"container/list"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

const (
	alpha = 3
	b     = 8 * IDBytes
	k     = 20
)

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID         ID
	SelfContact    Contact
	updateChannel  chan Contact
	findChannel    chan routingRequest
	getLastChannel chan routingRequest
	routingTable   []*KBucket
}

type routingRequest struct {
	NodeID          ID
	Count           int
	ResponseChannel interface{}
}

type probeResult struct {
	ProbeContact *list.Element
	Result       bool
}

func NewKademlia(laddr string) *Kademlia {
	// TODO: Initialize other state here as you add functionality.
	k := new(Kademlia)
	k.NodeID = NewRandomID()
	k.updateChannel = make(chan Contact, 10)
	k.routingTable = make([]*KBucket, b)
	for ii, _ := range k.routingTable {
		k.routingTable[ii] = NewKBucket()
	}
	k.findChannel = make(chan routingRequest)
	k.getLastChannel = make(chan routingRequest)

	// Set up RPC server
	// NOTE: KademliaCore is just a wrapper around Kademlia. This type includes
	// the RPC functions.
	rpc.Register(&KademliaCore{k})
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}
	// Run RPC server forever.
	go http.Serve(l, nil)

	// Add self contact
	hostname, port, _ := net.SplitHostPort(l.Addr().String())
	port_int, _ := strconv.Atoi(port)
	ipAddrStrings, err := net.LookupHost(hostname)
	var host net.IP
	for i := 0; i < len(ipAddrStrings); i++ {
		host = net.ParseIP(ipAddrStrings[i])
		if host.To4() != nil {
			break
		}
	}
	k.SelfContact = Contact{k.NodeID, host, uint16(port_int)}
	fmt.Println("My ID: " + k.NodeID.AsString())
	go k.handleUpdate()
	return k
}

func (k *Kademlia) handleUpdate() {
	responseChannel := make(chan probeResult, 10)
	defer close(responseChannel)
	for {
		select {
		// TODO: handle update request
		case c := <-k.updateChannel:
			fmt.Println("This is update IP: " + c.Host.String() + ":" + strconv.Itoa(int(c.Port)))
			if c.NodeID.Equals(k.NodeID) {
				fmt.Println("**update the self NodeID")
				break
			}
			fmt.Println("**begin to update non-self NodeID")
		// TODO: handle find request
		case find := <-k.findChannel:
			idx := k.NodeID.Xor(find.NodeID).PrefixLen()
			var ct *Contact
			if idx < b {
				ct, _ = k.routingTable[idx].FindContact(find.NodeID)
			} else {
				ct = nil
			}
			find.ResponseChannel.(chan *Contact) <- ct
		// TODO: handle get last reqeust
		case get := <-k.getLastChannel:
			fmt.Println("get: " + get.NodeID.AsString())
		// TODO: handle ping response
		case res := <-responseChannel:
			if res.Result {
				fmt.Println("result true")
			} else {
				fmt.Println("result false")
			}
		}
	}
}

type NotFoundError struct {
	id  ID
	msg string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%x %s", e.id, e.msg)
}

func (k *Kademlia) findContactFromKRoutingTable(nodeId ID) *Contact {
	resCh := make(chan *Contact)
	k.findChannel <- routingRequest{nodeId, 0, resCh}
	ct := <-resCh
	close(resCh)
	return ct
}

func (k *Kademlia) getLastContactFromRoutingTable(nodeId ID) (ret []Contact) {
	resCh := make(chan []Contact)
	k.getLastChannel <- routingRequest{nodeId, 0, resCh}
	ret = <-resCh
	close(resCh)
	return
}

func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Find contact with provided ID
	if nodeId == k.SelfContact.NodeID {
		return &k.SelfContact, nil
	}
	ct := k.findContactFromKRoutingTable(nodeId)
	if ct != nil {
		return ct, nil
	}
	return nil, &NotFoundError{nodeId, "Not found"}
}

func GetClient(host net.IP, port uint16) *rpc.Client {
	peerStr := host.String() + ":" + strconv.Itoa(int(port))
	fmt.Println("peerstr:" + peerStr)
	client, err := rpc.DialHTTP("tcp", peerStr)
	if err != nil {
		return nil
	}
	return client
}

func (k *Kademlia) internalPing(host net.IP, port uint16, update bool) (id ID, ok bool) {
	client := GetClient(host, port)
	ok = true
	if client == nil {
		ok = false
		return
	}
	pingReq := new(PingMessage)
	pingReq.MsgID = NewRandomID()
	var pong PongMessage
	err := client.Call("KademliaCore.Ping", pingReq, &pong)
	if err != nil {
		ok = false
		return
	}
	id = pong.Sender.NodeID
	if update {
		k.updateChannel <- pong.Sender
	}
	return
}

// This is the function to perform the RPC
func (k *Kademlia) DoPing(host net.IP, port uint16) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	//return "ERR: Not implemented"
	id, ok := k.internalPing(host, port, true)
	if ok {
		return id.AsString()
	}
	return "Failed to ping"
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	return "ERR: Not implemented"
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	return "ERR: Not implemented"
}

func (k *Kademlia) DoFindValue(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	return "ERR: Not implemented"
}

func (k *Kademlia) LocalFindValue(searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	return "ERR: Not implemented"
}

func (k *Kademlia) DoIterativeFindNode(id ID) string {
	// For project 2!
	return "ERR: Not implemented"
}
func (k *Kademlia) DoIterativeStore(key ID, value []byte) string {
	// For project 2!
	return "ERR: Not implemented"
}
func (k *Kademlia) DoIterativeFindValue(key ID) string {
	// For project 2!
	return "ERR: Not implemented"
}
