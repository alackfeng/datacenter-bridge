
# consul docker (using orbstack).
docker pull hashicorp/consul
docker tag hashicorp/consul consul

docker volume create consul-data
docker network create consul-net

## dc1
#### docker stop consul1
docker run -d --name=consul1 -p 12900:8500 -e CONSUL_BIND_INTERFACE=eth0 consul agent -server=true -bootstrap-expect 3 -data-dir=/tmp/consul -client=0.0.0.0 -ui

JOIN_IP="$(docker inspect -f '{{.NetworkSettings.IPAddress}}' consul1)"

docker run -d --name=consul2 -e CONSUL_BIND_INTERFACE=eth0 consul agent -server=true -client=0.0.0.0 -join=$JOIN_IP
docker run -d --name=consul3 -e CONSUL_BIND_INTERFACE=eth0 consul agent -server=true -client=0.0.0.0 -join=192.168.215.2
docker run -d --name=consul4 -e CONSUL_BIND_INTERFACE=eth0 consul agent -server=true -client=0.0.0.0 -join=192.168.215.2

docker exec -it consul1 consul members

## us
docker run -d --name consul5 -p 12900:8500 -e CONSUL_BIND_INTERFACE=eth0 consul agent -server -bootstrap-expect 2 -datacenter=us -client=0.0.0.0 -ui 
docker run -d --name consul6 consul agent -server -datacenter=us -client=0.0.0.0 -retry-join 192.168.215.2
docker run -d --name consul7 consul agent -server -datacenter=us -client=0.0.0.0 -retry-join 192.168.215.2

docker run -d --name consul8 -p 8500:8500 -e CONSUL_BIND_INTERFACE=eth0 consul agent -datacenter=us -client=0.0.0.0 -retry-join 192.168.215.2
docker exec consul5 consul members

docker exec consul5 consul join -wan 192.168.215.2

docker exec consul5 consul members
docker exec consul5 consul catalog datacenters
docker exec consul5 consul catalog nodes
docker exec consul5 consul members -wan
docker exec consul5 consul operator raft list-peers



## other.
cd /Users/taurus/taurus/code/hopeway/consul/services/web
docker build -t web-service .
docker run -d --name=web-service web-service

docker cp ./docs/consul/dcbridge_service.json consul8:/consul/config
docker exec -it consul8 consul reload

curl http://192.168.215.5:8500/v1/catalog/service/gw-dcb-service?dc=us
curl http://192.168.215.3:8500/v1/health/service/gw-dcb-service?dc=us
curl http://192.168.215.2:8500/v1/health/service/gw-dcb-service?dc=us

curl 192.168.215.5:8500/v1/catalog/nodes
curl localhost:8500/v1/catalog/nodes
curl localhost:8500/v1/health/service/gw-dcb-service?dc=us

cd /Users/taurus/taurus/code/hopeway/consul/services/config
docker run -d --name nginx -p 80:80 -v $(pwd)/nginx.ctmpl:/etc/nginx/nginx.ctmpl -v $(pwd)/consul-template:/usr/local/bin/consul-template nginx
​
consul-template -consul-addr=http://consul1:12900 -template "/etc/nginx/nginx.ctmpl:/etc/nginx/conf.d/default.conf:nginx -s reload"
​


​
