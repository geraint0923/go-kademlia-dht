package kademlia

import (
	"net"
	"strconv"
	"testing"
	"time"
)

var testPort uint16 = 3000

func StringToIpPort(laddr string) (ip net.IP, port uint16, err error) {
	hostString, portString, err := net.SplitHostPort(laddr)
	if err != nil {
		return
	}
	ipStr, err := net.LookupHost(hostString)
	if err != nil {
		return
	}
	for i := 0; i < len(ipStr); i++ {
		ip = net.ParseIP(ipStr[i])
		if ip.To4() != nil {
			break
		}
	}
	portInt, err := strconv.Atoi(portString)
	port = uint16(portInt)
	return
}

func TestPing(t *testing.T) {
	lport1 := testPort
	testPort++
	lport2 := testPort
	testPort++
	instance1 := NewKademlia("localhost:"+strconv.Itoa(int(lport1)), nil)
	instance2 := NewKademlia("localhost:"+strconv.Itoa(int(lport2)), nil)
	host2, port2, _ := StringToIpPort("localhost:" + strconv.Itoa(int(lport2)))
	instance1.DoPing(host2, port2)
	time.Sleep(30 * time.Millisecond)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	contact1, err := instance2.FindContact(instance1.NodeID)
	if err != nil {
		t.Error("Instance 1's contact not found in Instance 2's contact list")
		return
	}
	if contact1.NodeID != instance1.NodeID {
		t.Error("Instance 1 ID incorrectly stored in Instance 2's contact list")
	}
	if contact2.NodeID != instance2.NodeID {
		t.Error("Instance 2 ID incorrectly stored in Instance 1's contact list")
	}
	t.Log("TestPing done successfully!\n")
	return
}

func TestStore(t *testing.T) {
	lport1 := testPort
	testPort++
	lport2 := testPort
	testPort++
	instance1 := NewKademlia("localhost:"+strconv.Itoa(int(lport1)), nil)
	instance2 := NewKademlia("localhost:"+strconv.Itoa(int(lport2)), nil)
	host2, port2, _ := StringToIpPort("localhost:" + strconv.Itoa(int(lport2)))
	instance1.DoPing(host2, port2)
	time.Sleep(30 * time.Millisecond)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	contact1, err := instance2.FindContact(instance1.NodeID)
	if err != nil {
		t.Error("Instance 1's contact not found in Instance 2's contact list")
		return
	}
	if contact1.NodeID != instance1.NodeID {
		t.Error("Instance 1 ID incorrectly stored in Instance 2's contact list")
	}
	if contact2.NodeID != instance2.NodeID {
		t.Error("Instance 2 ID incorrectly stored in Instance 1's contact list")
	}
	randID := NewRandomID()
	randVal := NewRandomID().AsString()
	instance2.DoStore(contact1, randID, []byte(randVal))
	_, val1 := instance1.LocalFindValue(randID)
	if val1 == nil {
		t.Error("Instance 1 should have the stored value from Instance 2")
		return
	}
	if string(val1) != randVal {
		t.Error("Instance 1 doesn't have the right value for key: " + string(val1) + "!=" + randVal)
		return
	}
	_, val2 := instance2.LocalFindValue(randID)
	if val2 != nil {
		t.Error("Instance 2 should not have the value for the key")
		return
	}

	t.Log("TestStore done successfully!\n")
	return
}

func TestFindNode(t *testing.T) {
	t.Log("TestFindNode done successfully!\n")
	return
}

func TestFindValue(t *testing.T) {
	t.Log("TestFindValue done successfully!\n")
	return
}

func TestIterativeFindNode(t *testing.T) {
	t.Log("TestIterativeFindNode done successfully!\n")
	return
}

func TestIterativeFindValue(t *testing.T) {
	t.Log("TestIterativeFindValue done successfully!\n")
	return
}

func TestIterativeStore(t *testing.T) {
	t.Log("TestIterativeStore done successfully!\n")
	return
}
