package node

import (
	"log"
	"sync"
	"sync/atomic"
)

var NodeDB = NodeMap{Map: make(map[string]*Node)}

type NodeMap struct {
	Map  map[string]*Node `json:"node"`
	lock sync.RWMutex     `json:"-"`
}

type NodeMapItem struct {
	Key   string
	Value *Node
}

func (m *NodeMap) Set(key string, value *Node) {

	m.lock.Lock()
	defer m.lock.Unlock()

	m.Map[key] = value

}

func (m *NodeMap) Get(key string) (node *Node, exists bool) {

	m.lock.RLock()
	defer m.lock.RUnlock()

	node, exists = m.Map[key]
	return

}

func (m *NodeMap) Del(key string) {

	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.Map, key)

}

// Iter iterates over the items in a concurrent map
// Each item is sent over a channel, so that
// we can iterate over the map using the builtin range keyword
func (m *NodeMap) Iter() <-chan NodeMapItem {
	c := make(chan NodeMapItem)

	f := func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		for k, v := range m.Map {
			c <- NodeMapItem{k, v}
		}
		close(c)
	}
	go f()

	return c
}

// New a new node struct
// Args:
//   nodeRequest
// Return:
//   new node pointer
func NewNode(nodeRequest NodeRequest) *Node {

	if nodeRequest.IpAddress == "" ||
		nodeRequest.User == "" ||
		nodeRequest.Passwd == "" {
		log.Println("Error: node ip,user,password must specify")
		return nil
	}

	subnet := allocateSubnet()
	if subnet == "" {
		log.Println("Error, no subnet allocated")
		return nil
	}

	newNode := Node{
		Name:      nodeRequest.Name,
		IpAddress: nodeRequest.IpAddress,
		UserName:  nodeRequest.User,
		Passwd:    nodeRequest.Passwd,
		Role:      nodeRequest.Role,
		Status:    NodeStatusInit,
		State:     NodeStateEnable,
		PortMap:   make(map[int]string),
		Subnet:    subnet,
	}

	return &newNode
}

//Get node pointer by name
//Return nil if not existed
func GetNodeByName(nodeName string) *Node {

	myNode, exists := NodeDB.Get(nodeName)
	if exists == false {
		return nil
	} else {
		return myNode
	}

}

//Set node state(enable/disbale)
func (myNode *Node) SetState(state NodeState) {

	myNode.stateMutex.Lock()
	defer myNode.stateMutex.Unlock()

	myNode.State = state

}

//Get node state(enable/disable)
func (myNode *Node) GetState() NodeState {

	myNode.stateMutex.RLock()
	defer myNode.stateMutex.RUnlock()

	return myNode.State

}

//Set node status
func (myNode *Node) SetStatus(status NodeStatus) {

	myNode.statusMutex.Lock()
	defer myNode.statusMutex.Unlock()

	myNode.Status = status

}

//Get node status
func (myNode *Node) GetStatus() NodeStatus {

	myNode.statusMutex.RLock()
	defer myNode.statusMutex.RUnlock()

	return myNode.Status

}

//Reboot node
//Return nil if ok, otherwise error
func (myNode *Node) RebootNode() error {
	//TODO
	return nil
}

func (myNode *Node) ChangeCpuUsed(delta int32) {

	atomic.AddInt32(&myNode.CpuUsed, delta)
	return

}

func (myNode *Node) GetCpuUsed() (value int32) {

	value = atomic.LoadInt32(&myNode.CpuUsed)
	return

}

func (myNode *Node) ChangeMemUsed(delta int32) {

	atomic.AddInt32(&myNode.MemUsed, delta)
	return

}

func (myNode *Node) GetMemUsed() (value int32) {

	value = atomic.LoadInt32(&myNode.MemUsed)
	return

}

func (myNode *Node) ChangeDiskUsed(delta int32) {

	atomic.AddInt32(&myNode.DiskUsed, delta)
	return

}

func (myNode *Node) GetDiskUsed() (value int32) {

	value = atomic.LoadInt32(&myNode.DiskUsed)
	return

}

//Apply a available port, return 0 if not found
func (myNode *Node) ReservePort(destination string) int {

	myNode.portMutex.Lock()
	defer myNode.portMutex.Unlock()

	var port = NodePortRangeMin
	for ; port <= NodePortRangeMax; port++ {
		_, exists := myNode.PortMap[port]
		if exists {
			continue
		} else {
			myNode.PortMap[port] = destination
			break
		}
	}

	if port > NodePortRangeMax {
		return 0
	} else {
		return port
	}

}

//Return node port
func (myNode *Node) ReleasePort(port int) {

	myNode.portMutex.Lock()
	defer myNode.portMutex.Unlock()

	delete(myNode.PortMap, port)

}
