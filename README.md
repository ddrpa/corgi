# Corgi - 打印 HTTP 请求

```shell
$ ./corgi -h
usage: corgi [-h|--help] [-p|--port <integer>] [--max-printable-size <integer>]
             [--pretty] [--fetch "<value>"]

             Corgi HTTP Request Logger, version 1.1.0

Arguments:

  -h  --help                Print help information
  -p  --port                监听指定端口. Default: 8000
      --max-printable-size  请求体最大打印长度（0
                            表示不截断），JSON 和 URLEncoded
                            表单不受影响）. Default: 256
      --pretty              特定类型请求体输出美化
      --fetch               转发请求到指定地址
```

调试在线 HTTP Endpoint 可能需要为运行中的服务添加日志追踪点，重新打包和部署应用。
Corgi 可以通过代理接口在不修改代码的情况下展示请求信息。

可能有一些同学已经用过这个 Z shell 函数了，函数会打印指定端口收到的 HTTP 请求并返回 `200 OK`，在调试时非常有用（，
缺点大概是还没用 Bash 重写，而且无限循环的退出机制也没有认真研究过）。Corgi 是这个想法的扩充。

```shell
listen_port () {
	while true
	do
		{
			echo -e 'HTTP/1.1 200 OK\r\n'
		} | nc -l -v $1
		echo '\r\n'
	done
}
```

## AI 使用声明

项目使用 AI 辅助编码，用于 bug 查找和修复。

## HowTo

### 美化输出 `application/json` 和 `application/x-www-form-urlencoded` 类型的请求体

后者会按行列出键值对，其中的值会被 URL 解码。

```shell
$ ./corgi -p 8000 --pretty

2023/07/06 16:27:32 corgi is waiting on :8000
2023/07/06 16:27:38 POST /proxy?url=/iot/alipayApi/faceAuth/getAlipayUserInfo HTTP/1.1
RemoteAddr: [::1]:57382
Host: localhost:8000
cookie: Cookie_1=value
authorization: Bearer Igp5d444444444444444
user-agent: PostmanRuntime/7.32.3
accept: */*
accept-encoding: gzip, deflate, br
content-type: application/x-www-form-urlencoded
content-length: 98
postman-token: 9a00e0be-f921-4605-b2f3-b577c1e263c2
connection: keep-alive

payload={"username":"admin","password":"wecsnuigb43j@_f"}
method=PATCH
```

### 把请求转发到目标地址，并且将目标地址的响应返回给客户端

```shell
$ ./corgi -p 8000 --fetch localhost:8001 --pretty

2023/07/07 16:37:40 POST /proxy?url=/iot/alipayApi/faceAuth/getAlipayUserInfo HTTP/1.1
> RemoteAddr: [::1]:59214
> Host: localhost:8000
> user-agent: PostmanRuntime/7.32.3
> content-type: application/x-www-form-urlencoded
> cookie: Cookie_1=value
> content-length: 98
> authorization: Bearer Igp5d444444444444444
> accept: */*
> postman-token: 5dfc18b4-5902-4a15-94f5-53b23db764e7
> accept-encoding: gzip, deflate, br
> connection: keep-alive
> 
> payload={"username":"admin","password":"wecsnuigb43j@_f"}
> method=PATCH

2023/07/07 16:37:40 HTTP/1.1 200 200 OK
< date: Fri, 07 Jul 2023 08:37:40 GMT
< content-length: 9
< content-type: text/plain; charset=utf-8
< 
< woof woof
```
