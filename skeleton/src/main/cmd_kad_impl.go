package main

import (
	"fmt"
	"kademlia"
	"net"
	"strconv"
	"strings"
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
		if len(args) >= 1 {
			strList := strings.Split(args[0], ":")
			if len(strList) > 1 {
				// TODO: parse the host and the port
				port, err := strconv.Atoi(strList[1])
				if err == nil {
					success := kademlia.DoPing(k, net.ParseIP(strList[0]), uint16(port))
					if success {
						fmt.Println("command ping success")
					} else {
						fmt.Println("command ping failed")
					}
				} else {
					fmt.Println("Failed to parse port")
				}
			} else {
				// TODO: parse the ID
				curID, err := kademlia.FromString(args[0])
				if err == nil {
					// FIXME: handle whether the ID has showed up in the route table
					fmt.Println("Not implemented")
				}
			}
		} else {
			fmt.Println("usage: ping [host:port]")
		}
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
