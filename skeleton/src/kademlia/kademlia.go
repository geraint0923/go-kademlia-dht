package kademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

// Core Kademlia type. You can put whatever state you want in this.
type Kademlia struct {
	NodeID     ID
	MatchLevel int
	RouteTable [IDBytes*8 + 1]*KBucket
}

func NewKademlia() (k *Kademlia) {
	// TODO: Assign yourself a random ID and prepare other state here.
	k = new(Kademlia)
	k.NodeID = NewRandomID()
	return
}
