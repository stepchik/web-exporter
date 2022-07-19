package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Site struct {
	Name         string   `json:"name"`
	Site         string   `json:"site"`
	Method       string   `json:"method,omitempty"`
	Data         string   `json:"data,omitempty"`
	Contains     []string `json:"contains,omitempty"`
	ResponseCode int      `json:"response_code,omitempty"`
}

func doRequest(url string, method string, data string, waitTimeout time.Duration, response chan *fasthttp.Response) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)

	resp := fasthttp.AcquireResponse()

	if method == "POST" || method == "PUT" {
		req.Header.SetMethod(method)
		req.SetBodyString(data)
	}

	client := &fasthttp.Client{
		NoDefaultUserAgentHeader: true,
		TLSConfig:                &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:          2000,
		MaxIdleConnDuration:      waitTimeout,
		MaxConnDuration:          waitTimeout,
		ReadTimeout:              waitTimeout,
		WriteTimeout:             waitTimeout,
		MaxConnWaitTimeout:       waitTimeout,
	}

	_ = client.Do(req, resp)
	response <- resp
}

func checkSites(fileName string, waitTimeout time.Duration) string {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	var sites []Site
	err = json.Unmarshal(byteValue, &sites)
	if err != nil {
		log.Fatal(err)
	}
	result := ""

	if len(sites) > 0 {
		result += "# HELP web_sites Number 1 if site available\n"
		result += "# TYPE web_sites gauge\n"
	}

	for i := 0; i < len(sites); i++ {

		if sites[i].Method == "" {
			sites[i].Method = "GET"
		}

		if sites[i].ResponseCode == 0 {
			sites[i].ResponseCode = fasthttp.StatusOK
		}

		start := time.Now()
		response := make(chan *fasthttp.Response)
		go doRequest(sites[i].Site, sites[i].Method, sites[i].Data, waitTimeout, response)
		statusResult := 0

		resp := <-response
		if sites[i].ResponseCode == resp.StatusCode() {
			statusResult = 1
		}

		if statusResult == 1 {
			for _, v := range sites[i].Contains {
				if !strings.Contains(string(resp.Body()), v) {
					statusResult = 0
					break
				}
			}
		}
		elapsed := time.Since(start)
		result += "web_sites{name=\"" + sites[i].Name + "\",elapsed=\"" + fmt.Sprintf("%f", elapsed.Seconds()) + "\"} " + strconv.Itoa(statusResult) + "\n"
	}
	return result
}

func main() {
	listenAddr := flag.String("s", "0.0.0.0:5555", "Listen address. Example 0.0.0.0:5555")
	fileName := flag.String("f", "sites.json", "Local file")
	maxWaitResponseTime := flag.Duration("w", 5, "Request wait timeout")
	flag.Parse()

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Path()) == "/metrics" {
			result := checkSites(*fileName, *maxWaitResponseTime*time.Second)
			ctx.SetBodyString(result)
			return
		}
	}

	if err := fasthttp.ListenAndServe(*listenAddr, requestHandler); err != nil {
		log.Fatalf("error in ListenAndServe: %v", err)
	}
}
