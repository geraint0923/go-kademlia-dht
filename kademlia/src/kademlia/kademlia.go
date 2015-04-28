package kademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"bytes"
	"container/heap"
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
	B     = 8 * IDBytes
	K     = 20
)

type Storage interface {
	Get(key ID) ([]byte, bool)
	Put(key ID, value []byte) bool
}

// Kademlia type. You can put whatever state you need in this.
type Kademlia struct {
	NodeID         ID
	SelfContact    Contact
	updateChannel  chan Contact
	findChannel    chan routingRequest
	getLastChannel chan routingRequest
	routingTable   []*KBucket
	storage        Storage
}

type routingRequest struct {
	NodeID          ID
	Count           int
	ResponseChannel interface{}
}

type probeResult struct {
	TargetKBucket  *KBucket
	ProbeContact   *list.Element
	ReplaceContact *Contact
	Result         bool
}

func NewKademlia(laddr string) *Kademlia {
	// TODO: Initialize other state here as you add functionality.
	k := new(Kademlia)
	k.NodeID = NewRandomID()
	k.updateChannel = make(chan Contact, 10)
	k.routingTable = make([]*KBucket, B)
	for ii, _ := range k.routingTable {
		k.routingTable[ii] = NewKBucket()
	}
	k.findChannel = make(chan routingRequest)
	k.getLastChannel = make(chan routingRequest)
	k.storage = NewLocalStorage()

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

type ContactHeap struct {
	List   []Contact
	NodeID ID
}

func (h ContactHeap) Len() int {
	return len(h.List)
}

func (h ContactHeap) Less(ii, jj int) bool {
	dist_i := h.NodeID.Xor(h.List[ii].NodeID)
	dist_j := h.NodeID.Xor(h.List[jj].NodeID)
	for i := 0; i < IDBytes; i++ {
		for j := 7; j >= 0; j-- {
			bit_i := (dist_i[i] >> uint8(j)) & 0x1
			bit_j := (dist_j[i] >> uint8(j)) & 0x1
			if bit_i < bit_j {
				return true
			} else if bit_j < bit_i {
				return false
			}
		}
	}
	return false
}

func (h ContactHeap) Swap(i, j int) {
	h.List[i], h.List[j] = h.List[j], h.List[i]
}

func (h *ContactHeap) Pop() interface{} {
	old := h.List
	n := len(old)
	x := old[n-1]
	h.List = old[0 : n-1]
	return x
}

func (h *ContactHeap) Push(x interface{}) {
	h.List = append(h.List, x.(Contact))
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
			idx := k.NodeID.Xor(c.NodeID).PrefixLen()
			if idx < B {
				ct, _ := k.routingTable[idx].FindContact(c.NodeID)
				if ct != nil {
					k.routingTable[idx].MoveToBack(ct)
				} else {
					if k.routingTable[idx].Full() {
						head := k.routingTable[idx].Front()
						hc := head.Value.(Contact)
						go func() {
							_, ok := k.internalPing(hc.Host, hc.Port, false)
							responseChannel <- probeResult{k.routingTable[idx], head, &c, ok}
						}()
					} else {
						k.routingTable[idx].PushBack(c)
					}
				}

			}
			fmt.Println("**begin to update non-self NodeID")
		// TODO: handle find request
		case find := <-k.findChannel:
			idx := k.NodeID.Xor(find.NodeID).PrefixLen()
			var ct *Contact
			if idx < B {
				ele, _ := k.routingTable[idx].FindContact(find.NodeID)
				if ele != nil {
					tct := ele.Value.(Contact)
					ct = &tct
				}
			}
			find.ResponseChannel.(chan *Contact) <- ct
		// TODO: handle get last reqeust
		case get := <-k.getLastChannel:
			fmt.Println("get: " + get.NodeID.AsString())
			cl := []Contact{}
			idx := k.NodeID.Xor(get.NodeID).PrefixLen()
			if idx >= B {
				idx = B - 1
			}
			curCount := 0
			lidx := idx
			for lidx < B && get.Count > curCount {
				tl := k.routingTable[lidx].GetLast(K)
				cl = append(cl, tl...)
				curCount += len(tl)
				//fmt.Println(lidx)
				lidx += 1
			}
			lidx = idx - 1
			for lidx >= 0 && get.Count > curCount {
				tl := k.routingTable[lidx].GetLast(K)
				cl = append(cl, tl...)
				curCount += len(tl)
				//fmt.Println(lidx)
				lidx -= 1
			}
			cHeap := &ContactHeap{cl, get.NodeID}
			heap.Init(cHeap)
			respList := []Contact{}
			for get.Count > 0 && cHeap.Len() > 0 {
				respList = append(respList, heap.Pop(cHeap).(Contact))
				get.Count -= 1
			}
			get.ResponseChannel.(chan []Contact) <- respList
		// TODO: handle ping response
		case res := <-responseChannel:
			if res.Result {
				fmt.Println("result true")
				res.TargetKBucket.MoveToBack(res.ProbeContact)
			} else {
				fmt.Println("result false")
				res.TargetKBucket.Remove(res.ProbeContact)
				res.TargetKBucket.PushBack(res.ReplaceContact)
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
	k.getLastChannel <- routingRequest{nodeId, K, resCh}
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
	pingReq.Sender = k.SelfContact
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
		return host.String() + ":" + strconv.Itoa(int(port)) + " has NodeID: " + id.AsString()
	}
	return "Failed to ping"
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	//return "ERR: Not implemented"
	client := GetClient(contact.Host, contact.Port)
	if client == nil {
		return "Failed to connect to " + contact.NodeID.AsString()
	}
	req := new(StoreRequest)
	req.Sender = k.SelfContact
	req.MsgID = NewRandomID()
	req.Key = key
	req.Value = value
	var res StoreResult
	err := client.Call("KademliaCore.Store", req, &res)
	if err != nil {
		return "ERR: Store on " + contact.NodeID.AsString() + "(" + contact.Host.String() + ":" + strconv.Itoa(int(contact.Port)) + ") : " + err.Error()
	}
	if !res.MsgID.Equals(req.MsgID) || res.Err != nil {
		return "ERR: Remote Store on " + contact.NodeID.AsString() + "(" + contact.Host.String() + ":" + strconv.Itoa(int(contact.Port)) + ") : " + err.Error()
	}
	return "OK: " + contact.NodeID.AsString()
}

func (k *Kademlia) internalFindNode(contact *Contact, searchKey ID) (res FindNodeResult, ok bool) {
	client := GetClient(contact.Host, contact.Port)
	if client == nil {
		fmt.Println("Failed to connect to " + contact.NodeID.AsString())
		ok = false
		return
	}
	req := new(FindNodeRequest)
	req.Sender = k.SelfContact
	req.MsgID = NewRandomID()
	req.NodeID = searchKey
	err := client.Call("KademliaCore.FindNode", req, &res)
	if err != nil || !req.MsgID.Equals(res.MsgID) {
		fmt.Println("Call error when calling FindNode remotely: ", contact.NodeID.AsString())
		ok = false
		return
	}
	ok = true
	return
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	//return "ERR: Not implemented"
	res, ok := k.internalFindNode(contact, searchKey)
	if !ok {
		return "ERR: FindNode failed: " + searchKey.AsString()
	}
	var buffer bytes.Buffer
	for idx, val := range res.Nodes {
		buffer.WriteString("\n[" + strconv.Itoa(idx) + "] NodeID: " + val.NodeID.AsString() + " => " + val.Host.String() + ":" + strconv.Itoa(int(val.Port)))
	}
	return "OK: FindNode result =>" + buffer.String()
}

func (k *Kademlia) internalFindValue(contact *Contact, searchKey ID) (res FindValueResult, ok bool) {
	client := GetClient(contact.Host, contact.Port)
	if client == nil {
		fmt.Println("Failed to connect to " + contact.NodeID.AsString())
		ok = false
		return
	}
	req := new(FindValueRequest)
	req.Sender = k.SelfContact
	req.MsgID = NewRandomID()
	req.Key = searchKey
	err := client.Call("KademliaCore.FindValue", req, &res)
	if err != nil || !req.MsgID.Equals(res.MsgID) {
		fmt.Println("Call error when calling FindNode remotely: ", contact.NodeID.AsString())
		ok = false
		return
	}
	ok = true
	return
}

func (k *Kademlia) DoFindValue(contact *Contact, searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	//return "ERR: Not implemented"
	res, ok := k.internalFindValue(contact, searchKey)
	if !ok {
		return "ERR: FindValue failed: " + searchKey.AsString()
	}
	var buffer bytes.Buffer
	if res.Value != nil {
		buffer.WriteString(" Value = " + string(res.Value))
	} else {
		for idx, val := range res.Nodes {
			buffer.WriteString("\n[" + strconv.Itoa(idx) + "] NodeID: " + val.NodeID.AsString() + " => " + val.Host.String() + ":" + strconv.Itoa(int(val.Port)))
		}
	}
	return "OK: FindValue result =>" + buffer.String()
}

func (k *Kademlia) LocalFindValue(searchKey ID) string {
	// TODO: Implement
	// If all goes well, return "OK: <output>", otherwise print "ERR: <messsage>"
	//return "ERR: Not implemented"
	res, ok := k.storage.Get(searchKey)
	if ok {
		return "OK: " + searchKey.AsString() + "(" + string(res) + ")"
	}
	return "ERR: Key(" + searchKey.AsString() + ") not found"
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
