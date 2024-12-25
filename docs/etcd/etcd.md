
# etcd 部署方案.

## docker macos
#### https://github.com/etcd-io/etcd/releases
```

ETCD_VER=v3.5.17

rm -rf /tmp/etcd-data.tmp && mkdir -p /tmp/etcd-data.tmp && \
  docker rmi gcr.io/etcd-development/etcd:${ETCD_VER} || true && \
  docker run \
  -p 2379:2379 \
  -p 2380:2380 \
  --mount type=bind,source=/tmp/etcd-data.tmp,destination=/etcd-data \
  --name etcd-gcr-${ETCD_VER} \
  gcr.io/etcd-development/etcd:${ETCD_VER} \
  /usr/local/bin/etcd \
  --name s1 \
  --data-dir /etcd-data \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --initial-cluster s1=http://0.0.0.0:2380 \
  --initial-cluster-token tkn \
  --initial-cluster-state new \
  --log-level info \
  --logger zap \
  --log-outputs stderr

docker exec etcd-gcr-${ETCD_VER} /usr/local/bin/etcd --version
docker exec etcd-gcr-${ETCD_VER} /usr/local/bin/etcdctl version
docker exec etcd-gcr-${ETCD_VER} /usr/local/bin/etcdctl endpoint health
docker exec etcd-gcr-${ETCD_VER} /usr/local/bin/etcdctl put foo bar
docker exec etcd-gcr-${ETCD_VER} /usr/local/bin/etcdctl get foo

```

## docker-compose macos
```

docker network create etcd-net --subnet 172.25.0.0/16

mkdir -p /data/containers/etcd/{data,config}

cd /Users/taurus/taurus/code/hopeway/datacenter-bridge/docs/etcd/

cp config/etcd.conf.yml /data/containers/etcd/config/etcd.conf.yml
cp docker-compose.yml /data/containers/etcd/docker-compose.yml

cd /data/containers/etcd
docker compose up -d
docker compose ps -a

etcdctl --endpoints=127.0.0.1:2379 --write-out=table endpoint health
etcdctl --endpoints=127.0.0.1:2379 --write-out=table member list
etcdctl --endpoints=127.0.0.1:2379 put foo bar
etcdctl --endpoints=127.0.0.1:2379 get foo

```

## docker-compose-cluster macos
```
rm -rf /tmp/etcd/cluster

# node1 node2 node3
mkdir -p /tmp/etcd/cluster/etcd1/{data,conf}  
mkdir -p /tmp/etcd/cluster/etcd2/{data,conf}  
mkdir -p /tmp/etcd/cluster/etcd3/{data,conf}  

cd /Users/taurus/taurus/code/hopeway/datacenter-bridge/docs/etcd

cp cluster/config/etcd1.conf.yml /tmp/etcd/cluster/etcd1/conf/etcd.yml
cp cluster/config/etcd2.conf.yml /tmp/etcd/cluster/etcd2/conf/etcd.yml
cp cluster/config/etcd3.conf.yml /tmp/etcd/cluster/etcd3/conf/etcd.yml

docker compose -f docker-compose-cluster.yaml up -d
etcdctl --endpoints=127.0.0.1:23791 --write-out=table member list

docker compose -f docker-compose-cluster.yaml down

```