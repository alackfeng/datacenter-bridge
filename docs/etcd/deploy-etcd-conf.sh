#!/bin/bash
set -e

etcd_name=${1:-"etcd-s1"}
docker_container_dir=${2:-"/data/containers"}

# 创建基础目录
mkdir -p ${docker_container_dir}/etcd/{data,config}

# 创建 etcd 配置文件
function deploy_etcd_config(){
cat > ${docker_container_dir}/etcd/config/etcd.conf.yml <<-EOF
name: ${etcd_name}
data-dir: /var/etcd
listen-client-urls: http://0.0.0.0:2379
advertise-client-urls: http://0.0.0.0:2379
listen-peer-urls: http://0.0.0.0:2380
initial-advertise-peer-urls: http://0.0.0.0:2380
initial-cluster: etcd-s1=http://0.0.0.0:2380
initial-cluster-token: etcd-cluster
initial-cluster-state: new
logger: zap
log-level: info
#log-outputs: stderr

EOF
}

# 创建 docker-compose 文件
function deploy_compose_config(){
cat > ${docker_container_dir}/etcd/docker-compose.yml <<-EOF
version: '3'

services:
  etcd:
    container_name: ${etcd_name}
    image: quay.io/coreos/etcd:v3.5.12
    command: /usr/local/bin/etcd --config-file=/var/lib/etcd/conf/etcd.conf.yml
    volumes:
      - ${DOCKER_VOLUME_DIRECTORY:-.}/data:/var/etcd
      - ${DOCKER_VOLUME_DIRECTORY:-.}/config/etcd.conf.yml:/var/lib/etcd/conf/etcd.conf.yml
      - "/etc/localtime:/etc/localtime:ro"
    ports:
      - 2379:2379
      - 2380:2380
    restart: always

networks:
  default:
    name: etcd-tier
    driver: bridge

EOF
}

# 创建 etcd 服务
function deploy_etcd(){
  cd ${docker_container_dir}/etcd
  docker compose up -d
}

# 验证 etcd 服务
function check_etcd(){
  cd ${docker_container_dir}/etcd
  docker compose ps
}

echo -e "\033[1;32m 1.Deploy etcd config.\n \033[0m"
deploy_etcd_config

echo -e "\033[1;32m 2.Deploy docker compose config.\n \033[0m"
deploy_compose_config

echo -e "\033[1;32m 3.Deploy etcd service.\n \033[0m"
deploy_etcd

echo -e "\033[1;32m 4.Check etcd service status. \033[0m"
check_etcd