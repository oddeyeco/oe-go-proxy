package mainfiles

import (
	"bytes"
	"net"
	"net/http"
	"strconv"
	"time"
)

var pause = false

func postData(data string) {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}
	body := bytes.NewBufferString(data)
	req, _ := http.NewRequest(http.MethodPost, to.destinationURL, body)
	//req.Header.Add("Content-Type", "application/json")
	if to.clientAuth {
		req.SetBasicAuth(to.clientUser, to.clientPass)
	}

	req.Header.Add("Content-Length", strconv.Itoa(len(data)))
	resp, err := client.Do(req)
	if err != nil {
		//fmt.Println("Error:", err)
		to.queue <- data
		pause = true
		chocho <- true
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		to.queue <- data
		pause = true
		chocho <- true
		//fmt.Println(resp.Status, resp.ContentLength, resp.Request, resp.Header)
		return
	} else {
		pause = false
		chocho <- false
	}
}
