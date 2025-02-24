

# cfssl 证书生成.

sudo apt install golang-cfssl -y

## 创建证书配置并生成 CA.
#### cfssl print-defaults config > ca-config.json
cfssl gencert -initca ca-config.json | cfssljson -bare ca

## 给 Consul 节点生成证书.
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=server server-config.json | cfssljson -bare server

## 生成 Client 证书.
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=client client-config.json | cfssljson -bare client


## QA问题集合.
Q1. Failed to load config file: No "signing" field presentFailed to parse input: unexpected end of JSON input
A1. 在ca-config.json增加signing字段 =>
{
  "signing": {
    "default": {
      "expiry": "8760h"
    },
    "profiles": {
      "server": {
        "expiry": "8760h",
        "usages": ["signing", "key encipherment", "server auth"]
      },
      "client": {
        "expiry": "8760h",
        "usages": ["signing", "key encipherment", "client auth"]
      }
    }
  }
}

