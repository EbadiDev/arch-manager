## Arch Manager configuration

### Network Settings

#### Raw
```
{
  "transport" : "Raw",
  "acceptProxyProtocol": false,
  "flow": "xtls-rprx-vision",
  "header": {
    "type": "none"
  },
  "socketSettings" : {
    "useSocket" : false,
    "DomainStrategy": "asis",
    "tcpKeepAliveInterval": 0,
    "tcpUserTimeout": 0,
    "tcpMaxSeg": 0,
    "tcpWindowClamp": 0,
    "tcpKeepAliveIdle": 0,
    "tcpMptcp": false
  }
}
```
#### Raw + HTTP
```
{
  "transport" : "Raw",
  "acceptProxyProtocol": false,
  "header": {
    "type": "http",
    "request": {
      "path": "/arch",
      "headers": {
        "Host": ["www.baidu.com", "www.taobao.com", "www.cloudflare.com"]
      }
    }
  },
  "socketSettings" : {
    "useSocket" : false,
    "DomainStrategy": "asis",
    "tcpKeepAliveInterval": 0,
    "tcpUserTimeout": 0,
    "tcpMaxSeg": 0,
    "tcpWindowClamp": 0,
    "tcpKeepAliveIdle": 0,
    "tcpMptcp": false
  }
}

HttpHeaderObject

HTTP masquerading configuration must be configured on the corresponding inbound and outbound connections at the same time, and the content must be consistent.

{
  "type": "http",
  "request": {},
  "response": {}
}

    type: "http" 

Specify HTTP masquerade

    request: HTTPRequestObject 

HTTP Request

    response: HTTPResponseObject 

HTTP response
HTTPRequestObject

{
  "version": "1.1",
  "method": "GET",
  "path": ["/"],
  "headers": {
    "Host": ["www.baidu.com", "www.bing.com"],
    "User-Agent": [
      "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
      "Mozilla/5.0 (iPhone; CPU iPhone OS 10_0_2 like Mac OS X) AppleWebKit/601.1 (KHTML, like Gecko) CriOS/53.0.2785.109 Mobile/14A456 Safari/601.1.46"
    ],
    "Accept-Encoding": ["gzip, deflate"],
    "Connection": ["keep-alive"],
    "Pragma": "no-cache"
  }
}

    version: string 

HTTP version, the default value is "1.1"。

    method: string 

HTTP method, the default value is "GET"。

    path: [ string ] 

path, an array of strings. The default value is ["/"]When there are multiple values, a random value is selected for each request.

    headers: map{ string, [ string ]} 

HTTP header, a key-value pair, each key represents the name of an HTTP header, and the corresponding value is an array.

Each request will include all keys and randomly select a corresponding value. The default value is shown in the example above.
HTTPResponseObject

{
  "version": "1.1",
  "status": "200",
  "reason": "OK",
  "headers": {
    "Content-Type": ["application/octet-stream", "video/mpeg"],
    "Transfer-Encoding": ["chunked"],
    "Connection": ["keep-alive"],
    "Pragma": "no-cache"
  }
}

    version: string 

HTTP version, the default value is "1.1"。

    status: string 

HTTP status, the default value is "200"。

    reason: string 

HTTP status description, the default value is "OK"。

    headers: map {string, [ string ]} 

HTTP header, a key-value pair, each key represents the name of an HTTP header, and the corresponding value is an array.

Each request will include all keys and randomly select a corresponding value. The default value is shown in the example above. 

```
####  WS
```
{
  "transport": "ws",
  "acceptProxyProtocol": false,
  "path": "/arch?ed=2560",
  "host": "hk1.xyz.com",
  "heartbeatperiod": 30,
  "custom_host": "fakedomain.com",
  "socketSettings" : {
    "useSocket" : false,
    "DomainStrategy": "asis",
    "tcpKeepAliveInterval": 0,
    "tcpUserTimeout": 0,
    "tcpMaxSeg": 0,
    "tcpWindowClamp": 0,
    "tcpKeepAliveIdle": 0,
    "tcpMptcp": false
  }
}
```

####  GRPC
```
{
  "transport" : "grpc",
  "acceptProxyProtocol": false,
  "serviceName": "arch",
  "authority": "hk1.xyz.com",
  "multiMode": false,
  "user_agent": "custom user agent",
  "idle_timeout": 60,
  "health_check_timeout": 20,
  "permit_without_stream": false,
  "initial_windows_size": 0
  "socketSettings" : {
    "useSocket" : false,
    "DomainStrategy": "asis",
    "tcpKeepAliveInterval": 0,
    "tcpUserTimeout": 0,
    "tcpMaxSeg": 0,
    "tcpWindowClamp": 0,
    "tcpKeepAliveIdle": 0,
    "tcpMptcp": false
  }
}
```


####  KCP
```
{
  "transport" : "kcp",
  "acceptProxyProtocol": false,
  "mtu": 1350,
  "tti": 20,
  "uplinkCapacity": 5,
  "downlinkCapacity": 20,
  "congestion": false,
  "readBufferSize": 1,
  "writeBufferSize": 1,
  "congestion": false,
  "header": {
    "type": "none"
  },
  "seed": "password",
  "socketSettings" : {
    "useSocket" : false,
    "DomainStrategy": "asis",
    "tcpKeepAliveInterval": 0,
    "tcpUserTimeout": 0,
    "tcpMaxSeg": 0,
    "tcpWindowClamp": 0,
    "tcpKeepAliveIdle": 0,
    "tcpMptcp": false
  }
}

HeaderObject

{
  "type": "none",
  "domain": "example.com"
}

    type: string 

Disguise type, optional values ​​are:

    "none": The default value, no camouflage is performed, and the data sent is a data packet without characteristics.
    "srtp": Disguised as SRTP packets, they will be identified as video call data (such as FaceTime).
    "utp": Disguised as uTP data packets, they will be identified as BT download data.
    "wechat-video": Data packets disguised as WeChat video calls.
    "dtls": Disguised as a DTLS 1.2 packet.
    "wireguard": Disguised as WireGuard packets. (Not the real WireGuard protocol)
    "dns"：Some campus networks allow DNS queries without logging in. By adding a DNS header to KCP and disguising the traffic as a DNS request, you can bypass some campus network logins. 

    domain: string 

Match the camouflage type "dns"Use, you can fill in any domain name.

```

####  HTTPUPGRADE
```
{
  "transport" : "httpupgrade",
  "acceptProxyProtocol": false,
  "host": "hk1.xyz.com",
  "path": "/arch?ed=2560",
  "custom_host": "fakedomain.com",
  "socketSettings" : {
    "useSocket" : false,
    "DomainStrategy": "asis",
    "tcpKeepAliveInterval": 0,
    "tcpUserTimeout": 0,
    "tcpMaxSeg": 0,
    "tcpWindowClamp": 0,
    "tcpKeepAliveIdle": 0,
    "tcpMptcp": false
  }
}
```

####  XHTTP
```
{
  "transport" : "XHTTP",
  "acceptProxyProtocol": false,
  "host": "hk1.xyz.com",
  "custom_host": "fakedomain.com",
  "path": "/",
  "noSSEHeader": false,
  "noGRPCHeader": true,
  "mode": "auto",
  "socketSettings" : {
    "useSocket" : false,
    "DomainStrategy": "asis",
    "tcpKeepAliveInterval": 0,
    "tcpUserTimeout": 0,
    "tcpMaxSeg": 0,
    "tcpWindowClamp": 0,
    "tcpKeepAliveIdle": 0,
    "tcpMptcp": false
  }
}
```