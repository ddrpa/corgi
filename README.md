# Corgi - 打印 HTTP 请求

```shell
$ ./corgi -h
usage: corgi [-h|--help] [-p|--port <integer>] [-m|--max-printable-size
             <integer>] [--pretty]

             Corgi HTTP Request Logger, version 0.0.1

Arguments:

  -h  --help                Print help information
  -p  --port                监听指定端口. Default: 8000
  -m  --max-printable-size  请求体最大打印长度（0
                            表示不截断），JSON 和 URLEncoded
                            表单不受影响）. Default: 256
      --pretty              特定类型请求体输出美化
```

可能有一些同学已经用过这个 Z shell 函数了，函数会打印指定端口收到的 HTTP 请求并返回 `200 OK`，在调试时非常有用。
缺点大概是还没用 Bash 重写，而且无限循环的退出机制也没有认真研究过。

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

Corgi 做了同样的事情：

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
