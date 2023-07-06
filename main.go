package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	Version            = "1.0.0"
	DefaultPort        = 8000
	DefaultMaxBodySize = 256
)

var maxPrintableBodySize int
var prettyPrint bool

func main() {
	parser := argparse.NewParser("corgi", "Corgi HTTP Request Logger, version "+Version)
	listenOnPort := parser.Int("p", "port", &argparse.Options{Required: false, Default: DefaultPort, Help: "监听指定端口"})
	maxBodySize := parser.Int("m", "max-printable-size", &argparse.Options{Required: false, Default: DefaultMaxBodySize, Help: "请求体最大打印长度（0 表示不截断），JSON 和 URLEncoded 表单不受影响）"})
	pretty := parser.Flag("", "pretty", &argparse.Options{Required: false, Help: "特定类型请求体输出美化"})

	err := parser.Parse(os.Args)
	if err != nil {
		log.Fatal(parser.Usage(err))
	}

	listenAddr := ":" + strconv.Itoa(*listenOnPort)
	log.Println("Corgi is waiting on " + strconv.Itoa(*listenOnPort))
	maxPrintableBodySize = *maxBodySize
	prettyPrint = *pretty

	err = http.ListenAndServe(listenAddr, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Println(formatRequest(r) + "\n\n")
			w.WriteHeader(http.StatusOK)
		}))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func formatRequest(r *http.Request) string {
	var request []string
	request = append(request, fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto))
	request = append(request, fmt.Sprintf("RemoteAddr: %v", r.RemoteAddr))
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	for name, values := range r.Header {
		name = strings.ToLower(name)
		for _, value := range values {
			request = append(request, fmt.Sprintf("%v: %v", name, value))
		}
	}
	if r.Body != nil {
		switch r.Header.Get("Content-Type") {
		case "application/json":
			request = append(request, "\n"+prettyJSON(r.Body))
		case "application/x-www-form-urlencoded":
			err := r.ParseForm()
			if err != nil {
				request = append(request, "\n"+err.Error())
				break
			} else if prettyPrint {
				request = append(request, "\n"+prettyURLEncoded(r.PostForm))
			} else {
				request = append(request, "\n"+r.Form.Encode())
			}
		default:
			request = append(request, "\n"+prettyRaw(r.Body))
		}
	}
	return strings.Join(request, "\n")
}

func prettyJSON(body io.ReadCloser) string {
	b, err := io.ReadAll(body)
	if err != nil {
		return err.Error()
	}
	if b != nil {
		return string(b)
	} else {
		return ""
	}
}

func prettyURLEncoded(form url.Values) string {
	var content []string
	for k, vs := range form {
		for _, v := range vs {
			unescaped, err := url.QueryUnescape(v)
			if err != nil {
				content = append(content, fmt.Sprintf("%v=%v", k, v))
			} else {
				content = append(content, fmt.Sprintf("%v=%v", k, unescaped))
			}
		}
	}
	return strings.Join(content, "\n")
}

func prettyRaw(body io.ReadCloser) string {
	b, err := io.ReadAll(body)
	if err != nil {
		return err.Error()
	} else {
		rawBodyString := string(b)
		if maxPrintableBodySize != 0 && len(rawBodyString) > maxPrintableBodySize {
			return rawBodyString[:80] + "\n[request body truncated...]"
		} else {
			return rawBodyString
		}
	}
}
