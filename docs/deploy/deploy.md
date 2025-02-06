



# docker compose consul server or client on macos orbstack.
### https://www.cnblogs.com/wangguishe/p/16532531.html

ssh orb
cd /home/taurus/deploy

#### 生成gossip 加密密钥.
docker run -it --rm -v consul:/consul hashicorp/consul:latest consul keygen > consul/encrypt_key

cat consul/encrypt_key
iIz10PiR7dNmuAaUMeMqyYJbOxeUrUbKC2AcJic8SkE=

#### 为 RPC 加密生成 TLS 证书.
docker run -it --rm -v ./consul/certs:/consul/config/certs/ hashicorp/consul:latest sh
cd /consul/config/certs/
consul tls ca create -domain consul
consul tls cert create -server -dc cn-001 -domain consul
consul tls cert create -server -dc us-001 -domain consul

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
docker exec -it consul-server-1  consul acl bootstrap

b1a0164c-f0c7-b2e0-2278-219305926b63

#### 创建代理令牌策略.
docker cp consul/agent-policy.hcl consul-server-1:/
docker exec -it consul-server-1 sh
export CONSUL_HTTP_TOKEN=b1a0164c-f0c7-b2e0-2278-219305926b63
consul acl policy create -name "agent-token" -description "Agent Token Policy" -rules @agent-policy.hcl
consul acl token create -description "Agent Token" -policy-name "agent-token"

a9cf1f7b-ac56-5d5b-ce07-e2e3469779b7

CONSUL_HTTP_TOKEN=a9cf1f7b-ac56-5d5b-ce07-e2e3469779b7 consul members

curl --header "X-Consul-Token: 1f5dd4aa-96a5-f00a-0c4a-41997baf6043" http://127.0.0.1:8500/v1/agent/members
curl --header "X-Consul-Token: a9cf1f7b-ac56-5d5b-ce07-e2e3469779b7" http://consul-server-1:8500/v1/health/service/xxx


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


##
cd /home/taurus/deploy

cp -rf /Users/taurus/taurus/code/hopeway/datacenter-bridge/docs/deploy/consul_base/ .

docker compose -f consul_base/docker-compose.yml pull
docker compose -f consul_base/docker-compose.yml up -d