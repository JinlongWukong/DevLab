package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func HttpSendJsonData(uri string, method string, data []byte) error {
	client := &http.Client{}
	req, _ := http.NewRequest(method, uri, bytes.NewBuffer(data))
	req.Header.Set("Content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		log.Println("http response successfully")
	} else {
		return fmt.Errorf("unexpected status-code returned %v", resp.StatusCode)
	}

	return nil
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
