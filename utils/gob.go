package utils

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
)

func GobStoreToFile(filename string, data interface{}) error {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, buffer.Bytes(), 0600)
	if err != nil {
		return err
	}
	return nil
}

func GobLoadFromFile(filename string, data interface{}) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(raw)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(data)
	if err != nil {
		return err
	}
	return nil
}
