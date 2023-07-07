package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// 格式化 HTTP 请求
func formatRequest(r *http.Request) []string {
	var request []string
	request = append(request, fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto))
	request = append(request, fmt.Sprintf("RemoteAddr: %v", r.RemoteAddr))
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	request = append(request, formatHeader(r.Header)...)
	if r.ContentLength > 0 {
		request = append(request, "")
		request = append(request, formatRequestBody(r, r.Header.Get("Content-Type"))...)
	}
	return request
}

// 格式化 HTTP 响应
func formatResponse(r *http.Response) []string {
	var response []string
	response = append(response, fmt.Sprintf("%v %v %v", r.Proto, r.StatusCode, r.Status))
	response = append(response, formatHeader(r.Header)...)
	if r.ContentLength > 0 {
		response = append(response, "")
		response = append(response, prettyRaw(r.Body, int(r.ContentLength))...)
	}
	return response
}

// 格式化请求头 / 响应头
func formatHeader(header http.Header) []string {
	var headers []string
	for name, values := range header {
		name = strings.ToLower(name)
		for _, value := range values {
			headers = append(headers, fmt.Sprintf("%v: %v", name, value))
		}
	}
	return headers
}

// 格式化请求体，body 会被消耗，需要复用时在调取前先复制一份
func formatRequestBody(r *http.Request, contentType string) []string {
	switch contentType {
	case "application/json":
		return prettyJSON(r.Body)
	case "application/x-www-form-urlencoded":
		err := r.ParseForm()
		if err != nil {
			return []string{err.Error()}
		} else if prettyPrint {
			return prettyURLEncoded(r.PostForm)
		} else {
			return []string{r.PostForm.Encode()}
		}
	default:
		return prettyRaw(r.Body, int(r.ContentLength))
	}
}

func prettyJSON(body io.ReadCloser) []string {
	b, err := io.ReadAll(body)
	if err != nil {
		return []string{err.Error()}
	}
	if b != nil {
		return []string{string(b)}
	} else {
		return []string{"[empty body]"}
	}
}

func prettyURLEncoded(form url.Values) []string {
	var kvpair []string
	for k, vs := range form {
		for _, v := range vs {
			unescaped, err := url.QueryUnescape(v)
			if err != nil {
				kvpair = append(kvpair, fmt.Sprintf("%v=%v", k, v))
			} else {
				kvpair = append(kvpair, fmt.Sprintf("%v=%v", k, unescaped))
			}
		}
	}
	return kvpair
}

func prettyRaw(body io.ReadCloser, contentLength int) []string {
	if maxPrintableBodySize == 0 || contentLength < maxPrintableBodySize {
		b, err := io.ReadAll(body)
		if err != nil {
			return []string{err.Error()}
		} else {
			return []string{string(b)}
		}
	} else {
		b := make([]byte, maxPrintableBodySize)
		_, err := io.ReadFull(body, b)
		if err != nil {
			return []string{err.Error()}
		} else {
			return []string{string(b), "[request body truncated...]"}
		}
	}
}
