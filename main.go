package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/akamensky/argparse"
)

const (
	Version            = "1.1.2"
	DefaultPort        = 8000
	DefaultMaxBodySize = 256
)

var maxPrintableBodySize int
var prettyPrint bool
var fetchRemoteAddr string
var httpClient *http.Client

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

	// 复用 http.Client 以利用连接池
	httpClient = &http.Client{}

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
			if err != nil {
				responseLog = []string{err.Error()}
				log.Println(strings.Join(responseLog, "\n< ") + "\n")
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			defer remoteResponse.Body.Close()
			// 暂存响应体
			responseBodyBytes := makeResponseBodyReusable(remoteResponse)
			responseLog = formatResponse(remoteResponse)
			// 响应记录前添加方向符号
			log.Println(strings.Join(responseLog, "\n< ") + "\n")
			// 把响应转发给原始请求者（必须先设置 Header，再调用 WriteHeader）
			for name, values := range remoteResponse.Header {
				for _, value := range values {
					w.Header().Add(name, value)
				}
			}
			w.WriteHeader(remoteResponse.StatusCode)
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
	// 构建目标 URL
	targetURL := &url.URL{
		Scheme:   "http",
		Host:     fetchRemoteAddr,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}
	newRequest, err := http.NewRequest(r.Method, targetURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	// 复制请求头（避免直接赋值导致共享引用）
	for name, values := range r.Header {
		for _, value := range values {
			newRequest.Header.Add(name, value)
		}
	}
	newRequest.ContentLength = int64(len(body))
	return httpClient.Do(newRequest)
}
