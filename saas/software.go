package saas

import "log"

var backendMap = map[string]string{
	"jenkins": "container",
}

//Factory a new software
func NewSoftware(name string, softwareRequest SoftwareRequest) *Software {

	backend, existed := backendMap[softwareRequest.Kind]
	if existed == false {
		log.Printf("Error: SaaS kind %v not supported", softwareRequest.Kind)
		return nil
	}

	newSoftware := Software{
		Name:            name,
		Kind:            softwareRequest.Kind,
		Backend:         backend,
		Version:         softwareRequest.Version,
		CPU:             softwareRequest.CPU,
		Memory:          softwareRequest.Memory,
		PortMapping:     map[string]string{},
		AdditionalInfor: map[string]string{},
	}

	newSoftware.SetStatus(SoftwareStatusInit)
	return &newSoftware
}

//Set Software status
func (mySoftware *Software) SetStatus(status SoftwareStatus) {

	mySoftware.statusMutex.RLock()
	defer mySoftware.statusMutex.RUnlock()

	mySoftware.Status = status

}

//Get Software status
func (mySoftware *Software) GetStatus() SoftwareStatus {

	mySoftware.statusMutex.RLock()
	defer mySoftware.statusMutex.RUnlock()

	return mySoftware.Status

}
