



# docker compose consul server or client on macos orbstack.
### https://www.cnblogs.com/wangguishe/p/16532531.html

ssh orb
cd /home/taurus/deploy

#### 生成gossip 加密密钥.
docker run -it --rm -v consul:/consul hashicorp/consul:latest consul keygen > consul/encrypt_key

cat consul/encrypt_key
iIz10PiR7dNmuAaUMeMqyYJbOxeUrUbKC2AcJic8SkE=

#### 为 RPC 加密生成 TLS 证书.
docker run -it --rm -v ./consul/certs2:/consul/config/certs/ harbor.houwei-tech.com/hopeway/consul:latest sh
cd /consul/config/certs/
consul tls ca create -domain consul -days 36500
consul tls cert create -server -dc cn-001 -domain consul -days 36500 -additional-dnsname="*server*.us-001.consul"
consul tls cert create -server -dc us-001 -domain consul -days 36500 -additional-dnsname="*server*.cn-001.consul"

#### 检查配置文件.
docker run -it --rm -v ./consul/config/consul-server-1.json:/consul/config/consul-server-1.json -v ./consul/certs:/consul/config/certs -v ./consul/data:/consul/data hashicorp/consul:latest consul validate /consul/config/
docker run -it --rm -v ./consul/config/consul-server-2.json:/consul/config/consul-server-2.json  hashicorp/consul:latest consul validate /consul/config/
docker run -it --rm -v ./consul/config/consul-server-3.json:/consul/config/consul-server-3.json  hashicorp/consul:latest consul validate /consul/config/
docker run -it --rm -v ./consul/config/consul-client-1.json:/consul/config/consul-client-1.json  hashicorp/consul:latest consul validate /consul/config/

#### 设置consul权限.
docker run -it --rm -v ./consul/config/consul-server-1.json:/consul/config/consul-server-1.json  hashicorp/consul:latest id consul
chown -R 100:1000 ./consul/logs
chown -R 100:1000 ./consul/config
chown -R 100:1000 ./consul/certs
chown -R 100:1000 ./consul/data

chown -R 100:1000 ./consul/logs_us
chown -R 100:1000 ./consul/data_us

sudo chown -R taurus:taurus ./consul
orb pull /home/taurus/deploy/consul_20250206.zip

#### 创建 Consul 数据中心.

docker network create --driver bridge --subnet=192.168.100.0/24 --gateway=192.168.100.1 consul

docker compose -f consul/docker-compose.yml up -d
docker compose -f consul/docker-compose.yml up consul-server-1 -d
docker compose -f consul/docker-compose.yml up consul-server-2 -d
docker compose -f consul/docker-compose.yml up consul-server-3 -d
docker compose -f consul/docker-compose.yml up consul-client-1 -d

docker compose -f consul/docker-compose.yml down

docker compose -f consul/docker-compose_us.yml up -d
docker compose -f consul/docker-compose_us.yml up consul-us-server-1 -d


#### acl令牌.
#### 创建初始管理令牌.
docker exec -it consul-server-2 consul acl bootstrap

consul acl set-agent-token replication 7a26e1bd-fdb9-b649-4567-26c33dc616fb

consul acl token list
consul acl policy list

b1a0164c-f0c7-b2e0-2278-219305926b63

curl --header "X-Consul-Token: 7a26e1bd-fdb9-b649-4567-26c33dc616fb" http://localhost:8500/v1/agent/services

consul acl token create -description "Replication Token" -policy-name=dc-sync-policy -token 7a26e1bd-fdb9-b649-4567-26c33dc616fb
consul acl set-agent-token replication 0c1c402c-c475-6aa2-b00b-0a6fb0462b75

#### 创建代理令牌策略.
docker cp consul/agent-policy.hcl consul-server-1:/
docker exec -it consul-server-1 sh
export CONSUL_HTTP_TOKEN=7a26e1bd-fdb9-b649-4567-26c33dc616fb
consul acl policy create -name "agent-token" -description "Agent Token Policy" -rules @agent-policy.hcl
consul acl token create -description "Agent Token" -policy-name "agent-token"

a9cf1f7b-ac56-5d5b-ce07-e2e3469779b7

CONSUL_HTTP_TOKEN=4b39a760-fedc-bd99-15dd-ad3b982c734a consul members

curl --header "X-Consul-Token: 3b12280d-26c6-4b2e-9b6d-f96fdbee11de" http://127.0.0.1:8500/v1/agent/members
curl --header "X-Consul-Token: a9cf1f7b-ac56-5d5b-ce07-e2e3469779b7" http://cn-server-1:8500/v1/health/service/gw-dcb-service
curl --header "X-Consul-Token: a9cf1f7b-ac56-5d5b-ce07-e2e3469779b7" http://10.206.0.137:8500/v1/health/service/gw-dcb-service


http://10.206.0.137:8500
 


#### 清理.
docker stop $(docker ps -aq)
docker rm $(docker ps -aq)
docker network prune
docker volume prune
sudo systemctl restart docker


#### 查看.
open http://localhost:8500/

docker exec -it consul-server-1 consul members -token "b1a0164c-f0c7-b2e0-2278-219305926b63"
docker exec -it consul-server-1 consul operator raft list-peers -token "b1a0164c-f0c7-b2e0-2278-219305926b63"


curl --header "X-Consul-Token: 3b12280d-26c6-4b2e-9b6d-f96fdbee11de" http://175.27.223.155:48080/v1/agent/members | jq
curl --header "X-Consul-Token: 4b39a760-fedc-bd99-15dd-ad3b982c734a" http://175.27.223.155:48080/v1/agent/members | jq
curl --header "X-Consul-Token: 4b39a760-fedc-bd99-15dd-ad3b982c734a" "http://175.27.223.155:48080/v1/health/service/gw-dcb-service?dc=cn-001" | jq


sudo docker exec -it consul-server-1 consul members -token "4b39a760-fedc-bd99-15dd-ad3b982c734a"
sudo docker exec -it consul-server-1 consul members -token "4b39a760-fedc-bd99-15dd-ad3b982c734a" -wan 

sudo docker exec -it consul-server-1 consul info -token "4b39a760-fedc-bd99-15dd-ad3b982c734a"
sudo docker exec -it consul-server-1 consul operator raft list-peers -token "4b39a760-fedc-bd99-15dd-ad3b982c734a"


##
cd /home/taurus/deploy

cp -rf /Users/taurus/taurus/code/hopeway/datacenter-bridge/docs/deploy/consul_base/ .

docker compose -f consul_base/docker-compose.yml pull
docker compose -f consul_base/docker-compose.yml up -d



## ######################

consul acl bootstrap

AccessorID:       522a40a0-0241-0940-bcbe-579a094d57bc
SecretID:         dda15adb-ffea-7501-c1a5-ddccb2bf35ff
Description:      Bootstrap Token (Global Management)
Local:            false
Create Time:      2025-02-20 12:09:58.190452928 +0000 UTC
Policies:
   00000000-0000-0000-0000-000000000001 - global-management


export CONSUL_HTTP_TOKEN=dda15adb-ffea-7501-c1a5-ddccb2bf35ff


consul acl token list
consul acl policy list
consul members
consul operator raft list-peers

consul acl policy create -name agent -rules @agent-policy.hcl
consul acl token create -description "agent token" -policy-name agent

/ # consul acl token create -description "agent token" -policy-name agent
AccessorID:       99bf7d70-6494-90fd-8f35-4143d120f8f8
SecretID:         57afb059-ef1e-eb9c-7527-b3e8be430c06
Description:      agent token
Local:            false
Create Time:      2025-02-20 12:14:06.408718174 +0000 UTC
Policies:
   2e1432ab-fe1b-3114-a267-765267012574 - agent



consul acl policy create -name replication -rules @replication-policy.hcl
consul acl token create -description "replication token" -policy-name replication

AccessorID:       0d822adc-d575-54c7-7d0b-d892b0f4056a
SecretID:         5d7b0d49-b682-29ff-ccb8-38605459f303
Description:      replication token
Local:            false
Create Time:      2025-02-20 12:11:35.093883164 +0000 UTC
Policies:
   ec4ebc82-d891-9f4d-57db-4b319365061e - replication

/ #

consul acl set-agent-token replication 27dce24c-3cb0-45cc-7406-c02a4d5e9cd7
consul acl set-agent-token default 27dce24c-3cb0-45cc-7406-c02a4d5e9cd7



/ # consul acl token create -description "Agent Token" -policy-name=agent-policy
AccessorID:       64331d28-bcd5-f538-5264-e9a81d42f2ec
SecretID:         130e6dad-a917-6689-2c0d-ea0d7dc34959
Description:      Agent Token
Local:            false
Create Time:      2025-02-20 11:50:50.868429928 +0000 UTC
Policies:
   764adbca-d4b0-e8ee-5d23-dbc7c776f093 - agent-policy


curl --header "X-Consul-Token: 57afb059-ef1e-eb9c-7527-b3e8be430c06" "http://175.27.223.155:8500/v1/health/service/gw-dcb-service?dc=cn-001" | jq
curl --header "X-Consul-Token: 57afb059-ef1e-eb9c-7527-b3e8be430c06" "http://139.162.61.214:8500/v1/health/service/gw-dcb-service?dc=cn-001" | jq
curl "http://139.162.61.214:8500/v1/health/service/gw-dcb-service?dc=cn-001"
