#### TLS
```
{
  "serverName": "arch.dev",
  "rejectUnknownSni": false,
  "allowInsecure": false,
  "fingerprint": "chrome",
  "sni": "arch.dev",
  "curvepreferences": "X25519",
  "alpn": [
    "h2",
    "http/1.1"
  ],
  "serverNameToVerify" : ""
}
```
#### REALITY

`Generate Private and Public Keys :   arch-manager x25519`
or we can implement something in web interface to gen privatekey and publickey

```
{
  "show" : false,
  "dest": "www.cloudflare.com:443",
  "privatekey" : "yBaw532IIUNuQWDTncozoBaLJmcd1JZzvsHUgVPxMk8",
  "minclientver":"",
  "maxclientver":"",
  "maxtimediff":0,
  "proxyprotocol":0,
  "shortids" : [
    "6ba85179e30d4fc2"
  ],
  "serverNames": [
    "www.cloudflare.com"
  ],
  "fingerprint": "chrome",
  "spiderx": "",
  "publickey": "7xhH4b_VkliBxGulljcyPOH-bYUA2dl-XAdZAsfhk04"
}