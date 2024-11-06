package main

import (
	"bytes"
	"github.com/akamensky/argparse"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	Version            = "1.1.1"
	DefaultPort        = 8000
	DefaultMaxBodySize = 256
)

var maxPrintableBodySize int
var prettyPrint bool
var fetchRemoteAddr string

func main() {
	parser := argparse.NewParser("corgi", "Corgi HTTP Request Logger, version "+Version)
	listenOnPort := parser.Int("p", "port", &argparse.Options{Required: false, Default: DefaultPort, Help: "监听指定端口"})
	maxBodySize := parser.Int("", "max-printable-size", &argparse.Options{Required: false, Default: DefaultMaxBodySize, Help: "请求体最大打印长度（0 表示不截断），JSON 和 URLEncoded 表单不受影响）"})
	pretty := parser.Flag("", "pretty", &argparse.Options{Required: false, Help: "特定类型请求体输出美化"})
	fetch := parser.String("", "fetch", &argparse.Options{Required: false, Help: "转发请求到指定地址"})

	err := parser.Parse(os.Args)
	if err != nil {
		log.Fatal(parser.Usage(err))
	}

	listenAddr := ":" + strconv.Itoa(*listenOnPort)
	log.Println("Corgi is waiting on " + strconv.Itoa(*listenOnPort))
	maxPrintableBodySize = *maxBodySize
	prettyPrint = *pretty
	fetchRemoteAddr = *fetch

	var handler func(w http.ResponseWriter, r *http.Request)
	if fetchRemoteAddr != "" {
		log.Println("Corgi will grab and retrieve request to " + fetchRemoteAddr)
		handler = func(w http.ResponseWriter, originRequest *http.Request) {
			// 暂存请求体
			requestBodyBytes := makeRequestBodyReusable(originRequest)
			requestLog := formatRequest(originRequest)
			// 请求记录前添加方向符号
			log.Println(strings.Join(requestLog, "\n> ") + "\n")

			// 把请求转发到指定地址
			remoteResponse, err := throwOutRequest(originRequest, requestBodyBytes)
			var responseLog []string
			// 暂存响应体
			responseBodyBytes := makeResponseBodyReusable(remoteResponse)
			if err != nil {
				responseLog = []string{err.Error()}
			} else {
				responseLog = formatResponse(remoteResponse)
			}
			// 响应记录前添加方向符号
			log.Println(strings.Join(responseLog, "\n< ") + "\n")
			// 把响应转发给原始请求者
			w.WriteHeader(remoteResponse.StatusCode)
			for name, values := range remoteResponse.Header {
				for _, value := range values {
					w.Header().Add(name, value)
				}
			}
			_, _ = w.Write(responseBodyBytes)
		}
	} else {
		handler = func(w http.ResponseWriter, r *http.Request) {
			log.Println(strings.Join(formatRequest(r), "\n") + "\n")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("woof woof"))
		}
	}
	err = http.ListenAndServe(listenAddr, http.HandlerFunc(handler))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func makeRequestBodyReusable(r *http.Request) []byte {
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes
}

func makeResponseBodyReusable(r *http.Response) []byte {
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes
}

func throwOutRequest(r *http.Request, body []byte) (*http.Response, error) {
	cli := &http.Client{}
	newRequest, err := http.NewRequest(r.Method, r.URL.String(), io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		return nil, err
	}
	newRequest.Header = r.Header
	newRequest.URL.Host = fetchRemoteAddr
	newRequest.URL.Scheme = "http"
	// TODO 有时候不能正确计算 Content-Length？
	newRequest.ContentLength = int64(len(body))
	response, err := cli.Do(newRequest)
	if err != nil {
		return nil, err
	}
	return response, nil
}
