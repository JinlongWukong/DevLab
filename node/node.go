package node

import (
	"sync/atomic"
)

var Node_db = make(map[string]*Node)

// Add a new node
// Args:
//   nodeRequest
// Return:
//   new node pointer
func NewNode(nodeRequest NodeRequest) *Node {

	newNode := Node{
		Name:      nodeRequest.Name,
		IpAddress: nodeRequest.IpAddress,
		UserName:  nodeRequest.User,
		Passwd:    nodeRequest.Passwd,
		Role:      nodeRequest.Role,
		Status:    NodeStatusInit,
		State:     NodeStateEnable,
	}

	return &newNode
}

//Get node pointer by name
//Return nil if not existed
func GetNodeByName(nodeName string) *Node {

	myNode, exists := Node_db[nodeName]
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
