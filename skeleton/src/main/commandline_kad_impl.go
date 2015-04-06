package main

import (
	"fmt"
	"kademlia"
)

func toKademlia(data interface{}) (ret *kademlia.Kademlia) {
	ret = nil
	if val, success := data.(*kademlia.Kademlia); success {
		fmt.Println("success")
		ret = val
	}
	return
}

func TestHehe(cmd string, args []string, data interface{}) error {
	fmt.Println("Test Command =>" + cmd)
	return nil
}

func CmdDoWhoami(cmd string, args []string, data interface{}) error {
	k := toKademlia(data)
	if k != nil {
	}
	return nil
}

func CmdDoLocalFindValue(cmd string, args []string, data interface{}) error {
	return nil
}

func CmdDoGetContact(cmd string, args []string, data interface{}) error {
	return nil
}

func CmdDoIterativeStore(cmd string, args []string, data interface{}) error {
	return nil
}

func CmdDoIterativeFindNode(cmd string, args []string, data interface{}) error {
	return nil
}

func CmdDoIterativeFindValue(cmd string, args []string, data interface{}) error {
	return nil
}

func CmdDoPing(cmd string, args []string, data interface{}) error {
	return nil
}

func CmdDoStore(cmd string, args []string, data interface{}) error {
	return nil
}

func CmdDoFindNode(cmd string, args []string, data interface{}) error {
	return nil
}

func CmdDoFindValue(cmd string, args []string, data interface{}) error {
	return nil
}
