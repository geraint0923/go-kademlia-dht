package main

import (
	"fmt"
	"kademlia"
)

func toKademlia(data interface{}) (ret *kademlia.Kademlia) {
	ret = nil
	if val, success := data.(*kademlia.Kademlia); success {
		ret = val
	}
	return
}

func CmdDoWhoami(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
		fmt.Println("Your NodeID is " + k.NodeID.AsString())
	}
	return nil
}

func CmdDoLocalFindValue(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
	}
	return nil
}

func CmdDoGetContact(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
	}
	return nil
}

func CmdDoIterativeStore(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
	}
	return nil
}

func CmdDoIterativeFindNode(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
	}
	return nil
}

func CmdDoIterativeFindValue(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
	}
	return nil
}

func CmdDoPing(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
	}
	return nil
}

func CmdDoStore(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
	}
	return nil
}

func CmdDoFindNode(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
	}
	return nil
}

func CmdDoFindValue(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
	}
	return nil
}
