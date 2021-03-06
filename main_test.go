package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func init() {
	// create main
	go main()
}

func TestMain(t *testing.T) {
	var app = "www.example.com"
	fakeServer := httptest.NewServer(http.HandlerFunc(ServeHTTP))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}
	p, _ := strconv.Atoi(u.Port())
	var upstreamConfig = APP{app, []string{"/"}, []string{"GET"}, []Backend{Backend{u.Hostname(), p, 1}}}

	// create config
	jsonBytes, _ := json.Marshal(upstreamConfig)
	resp, err := http.Post(
		"http://127.0.0.1:12345/app",
		"application/json",
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		t.Errorf("request error, got: %s", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("should return 200, but got: %d", resp.StatusCode)
	}
	if len(breaker.balancers) == 0 || len(breaker.routers) == 0 || len(breaker.timelines) == 0 {
		t.Errorf("breaker should been settled, but got: %+v", breaker)
	}

	// bad body
	resp, err = http.Post(
		"http://127.0.0.1:12345/app",
		"application/json",
		bytes.NewBuffer([]byte{}),
	)
	if err != nil || resp.StatusCode != 400 {
		t.Errorf("should return 200, but got: %s with code: %d", err, resp.StatusCode)
	}

	// bad json
	resp, err = http.Post(
		"http://127.0.0.1:12345/app",
		"application/json",
		bytes.NewBuffer([]byte("{")),
	)
	if err != nil || resp.StatusCode != 400 {
		t.Errorf("should return 200, but got: %s with code: %d", err, resp.StatusCode)
	}

	// bad config
	upstreamConfig.Methods = append(upstreamConfig.Methods, "POST")
	jsonBytes, _ = json.Marshal(upstreamConfig)
	resp, err = http.Post(
		"http://127.0.0.1:12345/app",
		"application/json",
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil || resp.StatusCode != 400 {
		t.Errorf("should return 200, but got: %s with code: %d", err, resp.StatusCode)
	}
}

func TestInspectAPP(t *testing.T) {
	var app = "www.example.com"
	fakeServer := httptest.NewServer(http.HandlerFunc(ServeHTTP))
	defer fakeServer.Close()
	u, err := url.ParseRequestURI(fakeServer.URL)
	if err != nil {
		t.Errorf("failed to parse fakeServer address: %s", fakeServer.URL)
	}
	p, _ := strconv.Atoi(u.Port())
	var upstreamConfig = APP{app, []string{"/"}, []string{"GET"}, []Backend{Backend{u.Hostname(), p, 1}}}

	// create config
	jsonBytes, _ := json.Marshal(upstreamConfig)
	resp, err := http.Post(
		"http://127.0.0.1:12345/app",
		"application/json",
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		t.Errorf("request error, got: %s", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("should return 200, but got: %d", resp.StatusCode)
	}
	if len(breaker.balancers) == 0 || len(breaker.routers) == 0 || len(breaker.timelines) == 0 {
		t.Errorf("breaker should been settled, but got: %+v", breaker)
	}

	// start to test inspect handler
	resp, err = http.Get("http://127.0.0.1:12345/inspect/this")
	if err != nil {
		t.Errorf("request error, got: %s", err)
	}
	if resp.StatusCode != 404 {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("should return 404, but got: %d with body: %s", resp.StatusCode, body)
	}

	resp, err = http.Get("http://127.0.0.1:12345/inspect/" + app)
	if err != nil {
		t.Errorf("request error, got: %s", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("should return 200, but got: %d", resp.StatusCode)
	}
}
