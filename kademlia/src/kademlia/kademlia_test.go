package kademlia

import (
	"bytes"
	"container/heap"
	"net"
	"strconv"
	"testing"
	"time"
)

var testPort uint16 = 3000

const testAddr = "localhost"
const divNum = 3

type KademliaList []*Kademlia

func GenerateRandomIDList(num int) (ret []ID) {
	ret = make([]ID, num)
	for i := 0; i < num; i++ {
		ret[i] = NewRandomID()
	}
	return
}

func GenerateTreeIDList(num int) (ret []ID) {
	ret = make([]ID, num)
	ret[0] = NewRandomID()
	for i := 1; i < num; i++ {
		if i > 150 {
			ret[i] = NewRandomID()
		} else {
			curID := ret[i/divNum]
			curID[i/8] = curID[i/8] ^ (1 << uint8(7-(i%8)))
			ret[i] = curID
		}
	}
	return ret
}

func GenerateTestList(num int, idList []ID) (kRet KademliaList, cRet []Contact) {
	kRet = []*Kademlia{}
	cRet = []Contact{}
	for i := 0; i < num; i++ {
		laddr := testAddr + ":" + strconv.Itoa(int(testPort))
		testPort++
		var k *Kademlia
		if idList != nil && i < len(idList) {
			k = NewKademlia(laddr, &idList[i])
		} else {
			k = NewKademlia(laddr, nil)
		}
		cRet = append(cRet, k.SelfContact)
		kRet = append(kRet, k)
	}
	return
}

func (ks KademliaList) ConnectTo(k1, k2 int) {
	ks[k1].DoPing(ks[k2].SelfContact.Host, ks[k2].SelfContact.Port)
}

func SortContact(input []Contact, key ID) (ret []Contact) {
	cHeap := &ContactHeap{input, key}
	heap.Init(cHeap)
	ret = []Contact{}
	for cHeap.Len() > 0 {
		ret = append(ret, heap.Pop(cHeap).(Contact))
	}
	return
}

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

func CompareContactList(l1, l2 []Contact) string {
	var buffer bytes.Buffer
	ll1 := len(l1)
	ll2 := len(l2)
	ll := ll1
	if ll2 > ll1 {
		ll = ll2
	}
	for i := 0; i < ll; i++ {
		buffer.WriteString("\n")
		if i < ll1 {
			buffer.WriteString(l1[i].NodeID.AsString())
		} else {
			buffer.WriteString("                    ")
		}
		buffer.WriteString("      ")
		if i < ll2 {
			buffer.WriteString(l2[i].NodeID.AsString())
		}
	}
	return buffer.String()
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
	kNum := K - 3
	testIdx := kNum/3 + 1
	kList, cList := GenerateTestList(kNum, nil)
	for i := 1; i < kNum; i++ {
		kList.ConnectTo(i, 0)
	}
	// wait for the completion of DoPing
	time.Sleep(100 * time.Millisecond)
	_, ret := kList[testIdx].DoFindNode(&kList[0].SelfContact, kList[testIdx].SelfContact.NodeID)
	if ret == nil {
		t.Error("The return of DoFindNode is nil!")
		return
	}
	sortedList := SortContact(cList[1:], kList[testIdx].SelfContact.NodeID)[1 : kNum-2+1]
	if len(ret) < len(sortedList) {
		t.Error("The number of returned contacts is less than " + strconv.Itoa(len(sortedList)) + ": " + strconv.Itoa(len(ret)))
		return
	}
	ret = SortContact(ret, kList[testIdx].SelfContact.NodeID)
	for idx := range sortedList {
		if !ret[idx].NodeID.Equals(sortedList[idx].NodeID) {
			t.Error(strconv.Itoa(idx) + " => NodeID not equal: " + ret[idx].NodeID.AsString() + "!=" + sortedList[idx].NodeID.AsString())
			t.Error(CompareContactList(sortedList, ret))
			t.Error("Source NodeID => " + kList[testIdx].SelfContact.NodeID.AsString())
			t.Error("Dest NodeID => " + kList[0].SelfContact.NodeID.AsString())
			return
		}
	}
	t.Log("TestFindNode done successfully!\n")
	return
}

func TestFindNodeLargeAndKBucket(t *testing.T) {
	kNum := K * 5
	testIdx := kNum/5*4 + 1
	idList := GenerateRandomIDList(kNum)
	// the first 2*K contacts should be in the same KBucket
	// however, only the first K contacts are able to reamin
	for i := 0; i < K*2; i++ {
		idList[i][0] = 0
		idList[i][1] = 0xef
	}
	idList[0][1] = 0xff
	for i := K * 2; i < kNum; i++ {
		idList[i][0] = 1
	}
	kList, cList := GenerateTestList(kNum, idList)
	for i := 1; i < kNum; i++ {
		kList.ConnectTo(i, 0)
		if i < K*2 {
			// make sure that the contact has hit the kbuckets
			time.Sleep(3 * time.Millisecond)
		}
	}
	// wait for the completion of DoPing
	time.Sleep(100 * time.Millisecond)
	_, ret := kList[testIdx].DoFindNode(&kList[0].SelfContact, kList[1].SelfContact.NodeID)
	if ret == nil {
		t.Error("The return of DoFindNode is nil!")
		return
	}
	// only the first K contacts(except the one indexed zero) remain in KBucket
	sortedList := SortContact(cList[1:K+1], kList[1].SelfContact.NodeID)
	if len(ret) < K {
		t.Error("The number of returned contacts is less than " + strconv.Itoa(K) + ": " + strconv.Itoa(len(ret)))
		return
	}
	ret = SortContact(ret, kList[1].SelfContact.NodeID)
	for idx := range sortedList {
		if !ret[idx].NodeID.Equals(sortedList[idx].NodeID) {
			t.Error(strconv.Itoa(idx) + " => NodeID not equal: " + ret[idx].NodeID.AsString() + "!=" + sortedList[idx].NodeID.AsString())
			t.Error(CompareContactList(sortedList, ret))
			t.Error("Source NodeID => " + kList[testIdx].SelfContact.NodeID.AsString())
			t.Error("Dest NodeID => " + kList[0].SelfContact.NodeID.AsString())
			return
		}
	}
	t.Log("TestFindNode done successfully!\n")
	return
}

func TestFindValue(t *testing.T) {
	t.Log("TestFindValue done successfully!\n")
	return
}

func TestIterativeFindNode(t *testing.T) {
	kNum := 120
	targetIdx := kNum - 23
	treeList := GenerateTreeIDList(kNum)
	kList, _ := GenerateTestList(kNum, treeList)
	for i := 1; i < kNum; i++ {
		kList.ConnectTo(i, i/divNum)
	}
	time.Sleep(100 * time.Millisecond)
	searchKey := kList[targetIdx].SelfContact.NodeID
	searchKey[IDBytes-1] = 0
	_, res := kList[0].DoIterativeFindNode(searchKey)
	res = SortContact(res, searchKey)
	if !res[0].NodeID.Equals(kList[targetIdx].SelfContact.NodeID) {
		t.Error("Search result doesn't match: " + res[0].NodeID.AsString() + "!=" + kList[targetIdx].SelfContact.NodeID.AsString())
	}
	t.Log("TestIterativeFindNode done successfully!\n")
	return
}

func TestIterativeFindValue(t *testing.T) {
	kNum := 120
	targetIdx := kNum - 23
	treeList := GenerateTreeIDList(kNum)
	kList, _ := GenerateTestList(kNum, treeList)
	for i := 1; i < kNum; i++ {
		kList.ConnectTo(i, i/divNum)
	}
	time.Sleep(100 * time.Millisecond)
	searchKey := kList[targetIdx].SelfContact.NodeID
	searchKey[IDBytes-1] = 0
	randValue := []byte(NewRandomID().AsString())
	kList[targetIdx/divNum].DoStore(&kList[targetIdx].SelfContact, searchKey, randValue)
	time.Sleep(3 * time.Millisecond)
	_, retVal := kList[targetIdx].LocalFindValue(searchKey)
	if retVal == nil {
		t.Error("The target node should have the key/value pair")
		return
	}
	if string(retVal) != string(randValue) {
		t.Error("The stored value should equal to each other")
		return
	}
	_, res, _ := kList[0].DoIterativeFindValue(searchKey)
	if res == nil {
		t.Error("The coressponding value should be found")
		return
	}
	if string(res) != string(randValue) {
		t.Error("Search result doesn't match: " + string(res) + "!=" + string(randValue))
	}
	t.Log("TestIterativeFindValue done successfully!\n")
	return
}

func TestIterativeStore(t *testing.T) {
	kNum := 120
	targetIdx := kNum - 23
	treeList := GenerateTreeIDList(kNum)
	kList, _ := GenerateTestList(kNum, treeList)
	for i := 1; i < kNum; i++ {
		kList.ConnectTo(i, i/divNum)
	}
	time.Sleep(100 * time.Millisecond)
	searchKey := kList[targetIdx].SelfContact.NodeID
	searchKey[IDBytes-1] = 0
	randValue := []byte(NewRandomID().AsString())
	// do the iterativeStore
	_, _ = kList[0].DoIterativeStore(searchKey, randValue)
	// retrive the value from target node
	_, retVal := kList[targetIdx].LocalFindValue(searchKey)
	if retVal == nil {
		t.Error("The target node should have the key/value pair")
		return
	}
	if string(retVal) != string(randValue) {
		t.Error("The stored value should equal to each other")
		return
	}
	t.Log("TestIterativeStore done successfully!\n")
	return
}
