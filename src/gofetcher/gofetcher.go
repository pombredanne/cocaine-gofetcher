package main

import (
	"cocaine"
	"codec"
	"net/http"
	"io/ioutil"
	"strconv"
	"fmt"
)

var (
	logger *cocaine.Logger
)

const (
	URL = iota
	BODY
	TIMEOUT
	COOKIES
	HEADERS
	FOLLOW_REDIRECTS
)

const (
	DefaultTimeout = 5000
)

type Request struct {
	method    string
	url       string
	body	[]byte
	timeout	int64
	cookies	map[string]string
	headers map[string][]string
}

type Response struct {
	httpResponse	*http.Response
	body	[]byte
	header	http.Header
}



func performRequest(request *Request) (*Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(request.method, request.url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response := &Response{httpResponse: resp, body: body, header: resp.Header}
	return response, nil

}

func tranformHeader(header http.Header) (hdr [][2]string) {
	for headerName, headerValues := range header {
		for _, headerValue := range headerValues {
			hdr = append(hdr, [2]string{headerName, headerValue})
		}
	}
	return hdr
}

func httpResponse(response *cocaine.Response, statusCode int, data interface{} , headers [][2]string) {
	response.Write(cocaine.WriteHead(statusCode, headers))
	response.Write(data)
	response.Close()
}

func httpGet(request *cocaine.Request, response *cocaine.Response) {
	var (
		timeout int64 = DefaultTimeout
	)
	req := cocaine.UnpackProxyRequest(<-request.Read())
	url := req.FormValue("url")
	if url == "" {
		url = "http://yandex.ru"
	}
	timeout_arg := req.FormValue("timeout")
	if timeout_arg != "" {
		tout, _ := strconv.Atoi(timeout_arg)
		timeout = int64(tout)
	}
	httpRequest := Request{method:"GET", url:url, timeout:timeout}
	resp, err := performRequest(&httpRequest)
	if err != nil {
		httpResponse(response, 500, err.Error(), [][2]string{{"Content-Type", "text/html"}})
	} else {
		httpResponse(response, 200, resp.body, tranformHeader(resp.header))
	}

}

func Get(request *cocaine.Request, response *cocaine.Response){
	requestBody := <- request.Read()
	var (
		mh codec.MsgpackHandle
		h = &mh
	)
	var res []interface{}
	codec.NewDecoderBytes(requestBody, h).Decode(&res)
	url := string(res[0].([]byte))
	httpRequest := Request{method:"GET", url:url}
	resp, err := performRequest(&httpRequest)
	if err != nil {
		response.Write([]interface{}{false, err.Error(), map[string][]string{}})
	} else{
		response.Write([]interface{}{true, fmt.Sprintf("%s", res[1].(int)), resp.header})
	}
	response.Close()
}

func main(){
	logger = cocaine.NewLogger()
	binds := map[string]cocaine.EventHandler{
		"get": Get,
		"httpget":      httpGet,

	}
	Worker := cocaine.NewWorker()
	Worker.Loop(binds)
}

func main_debug(){
	resp, err := performRequest(&Request{method:"GET", url:"http://yandex.ru", timeout:10})
	if err != nil {
		fmt.Println(err)

	} else {
		fmt.Println(string(resp.body))

	}
}
