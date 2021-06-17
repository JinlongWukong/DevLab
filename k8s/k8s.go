package k8s

import (
	"log"
	"time"
)

// New a k8s struct
// Args:
//   k8sRequest
// Return:
//   new k8s pointer
func NewK8s(name string, k8sRequest K8sRequest) *K8S {

	if name == "" ||
		k8sRequest.Version == "" ||
		k8sRequest.Type == "" {
		log.Println("Error: k8s name version, type must specify")
		return nil
	}

	newK8S := K8S{
		Name:             name,
		Type:             k8sRequest.Type,
		Version:          k8sRequest.Version,
		NumOfContronller: k8sRequest.NumOfContronller,
		NumOfWorker:      k8sRequest.NumOfWorker,
		Lifetime:         time.Duration(k8sRequest.Duration),
		Status:           K8sStatusInit,
	}

	return &newK8S
}

//Set k8s status
func (myK8s *K8S) SetStatus(status K8sStatus) {

	myK8s.Lock()
	defer myK8s.Unlock()

	myK8s.Status = status

}

//Get k8s status
func (myK8s *K8S) GetStatus() K8sStatus {

	myK8s.Lock()
	defer myK8s.Unlock()

	return myK8s.Status

}
