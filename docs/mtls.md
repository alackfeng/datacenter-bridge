# 使用 mTLS

## 生成证书
cd /Users/taurus/taurus/code/hopeway/datacenter-bridge/docs/mtls

### 1. 生成 CA 证书
```bash
#### 创建根证书私钥: password: 123123
openssl genrsa -aes256 -out root.key 2048
#### 创建根证书请求文件
openssl req -new -key root.key -out root.csr -utf8 -subj '/C=CN/ST=Guangdong/L=Guangzhou/O=bytezero.org/OU=bytezero/CN=www.bytezero.org/emailAddress=admin@bytezero.org'
#### 创建根证书
openssl x509 -req -in root.csr -signkey root.key -CAcreateserial  -out root.crt -days 36500 -sha256

```

### 2. 生成服务器证书
```bash
openssl genrsa -aes256 -out server/server.key 2048
openssl req -new -key server/server.key -out server/server.csr -utf8 -subj '/C=CN/ST=Guangdong/L=Guangzhou/O=bytezero.org/OU=bytezero/CN=www.bytezero.org/emailAddress=admin@bytezero.org'
openssl x509 -req -in server/server.csr -CA root.crt -CAkey root.key -CAcreateserial -out server/server.crt -days 365

```

### 3. 生成客户端证书
```bash
openssl genrsa -aes256 -out client/client.key 2048
openssl req -new -key client/client.key -out client/client.csr -utf8 -subj '/C=CN/ST=Guangdong/L=Guangzhou/O=bytezero.org/OU=bytezero/CN=www.bytezero.org/emailAddress=admin@bytezero.org'
openssl x509 -req -in client/client.csr -CA root.crt -CAkey root.key -CAcreateserial -out client/client.crt -days 365

```

### 4. 证书格式转换
```bash
#### 根证书转Java truststore JKS
keytool -import -file root.crt -alias rootCA -keystore truststore.jks
#### 服务端证书转PKCS#12
openssl pkcs12 -export -in server/server.crt -inkey server/server.key -certfile server/server.crt -out server/keystore.p12
#### 证书转PKCS#12
openssl pkcs12 -export -in server/server.crt -inkey server/server.key -out server/keystore.p12
openssl pkcs12 -export -clcerts -in client/client.crt -inkey client/client.key -out client/client.p12
#### PKCS#12转Java JKS
keytool -importkeystore -srckeystore keystore.p12 -srcstoretype pkcs12 -destkeystore keystore.jks -deststoretype JKS


```