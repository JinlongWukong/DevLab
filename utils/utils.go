package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

func HttpSendJsonData(uri string, method string, data []byte) (error, []byte) {

	client := &http.Client{}
	req, _ := http.NewRequest(method, uri, bytes.NewBuffer(data))
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		resp_body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("http response success with messages: %v", string(resp_body))
		return nil, resp_body
	} else {
		return fmt.Errorf("unexpected status-code returned %v", resp.StatusCode), nil
	}
}

func HttpSendJsonDataWithAuthBearer(uri, method, token string, data []byte) (error, []byte) {

	client := &http.Client{}
	req, _ := http.NewRequest(method, uri, bytes.NewBuffer(data))
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		resp_body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("http response success with messages: %v", string(resp_body))
		return nil, resp_body
	} else {
		return fmt.Errorf("unexpected status-code returned %v", resp.StatusCode), nil
	}
}

func HttpGetJsonData(uri string, query map[string]string) (error, []byte) {

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Add("Accept", "application/json")
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err, nil
	}

	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		resp_body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("http response success with messages: %v", string(resp_body))
		return nil, resp_body
	} else {
		return fmt.Errorf("unexpected status-code returned %v", resp.StatusCode), nil
	}
}

// Save a representation of v to the file at path
func WriteJsonFile(path string, v interface{}) error {

	file, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, file, 0644)
	if err != nil {
		return err
	}

	return err
}

// Load Json data from file path
func ReadJsonFile(path string) ([]byte, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Read json file failed: %v", err)
		return nil, err
	}

	return data, nil
}

// Random string generator
func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// Equal tells whether a and b contain the same elements.
func EqualStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
