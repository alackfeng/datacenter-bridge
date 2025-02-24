
# ansible 运维工具使用手册.

ssh orb

## 安装 ubuntu 24.04 TLS.
sudo apt install -y ansible

sudo mkdir -p /etc/ansible

## 配置目标服务器
#### Ansible 通过 SSH 连接到目标服务器
ssh-keygen -t rsa -b 2048
ssh-copy-id user@remote_host

## 使用 Ansible 部署 Consul
规划 Consul 集群架构

主机名	角色	IP 地址
server1	Consul Server	192.168.1.10
server2	Consul Server	192.168.1.11
server3	Consul Server	192.168.1.12
client1	Consul Client	192.168.1.13
client2	Consul Client	192.168.1.14

### 配置 inventory 文件
sudo vi /etc/ansible/hosts =>
[web_servers]
175.27.170.190 ansible_user=root
146.56.205.108 ansible_user=root

ansible -m ping all


## ansible galaxy roles
ansible-galaxy --version 
ls -ltr /home/taurus/.ansible

#### ls -ltr /home/taurus/.ansible/roles/geerlingguy.docker
ansible-galaxy install geerlingguy.docker

ansible-galaxy install geerlingguy.nginx

#### 自定义role
ansible-galaxy init roles/nginx


## 执行 Ansible Playbook
cd /home/taurus/datacenter-bridge/docs/ansible
sudo vi deploy-consul.yml =>




ansible-playbook deploy-consul.yml


## Ansible Inventory 中的所有主机.
ansible all --list-hosts
ansible consul_cluster --list-hosts
ansible -m setup cn-server-1 |grep "ansible_default_ipv4"

ansible-inventory --list
ansible-inventory --list -y
ansible-inventory --host cn-server-3



## centos install docker

ansible docker_nodes -m shell -a "sudo yum update -y" -v
ansible docker_nodes -m shell -a "sudo yum install -y yum-utils device-mapper-persistent-data lvm2" -v

ansible docker_nodes -m shell -a " sudo mv /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.backup " -v

ansible docker_nodes -m shell -a " sudo curl -o /etc/yum.repos.d/CentOS-Base.repo http://mirrors.aliyun.com/repo/Centos-7.repo " -v
ansible docker_nodes -m shell -a " curl -o /home/fengyue/CentOS-Base.repo http://mirrors.aliyun.com/repo/Centos-7.repo && sudo mv /home/fengyue/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo " -v

ansible docker_nodes -m shell -a " sudo yum clean all " -v
ansible docker_nodes -m shell -a " sudo yum makecache " -v

ansible docker_nodes -m shell -a " wget -O /home/fengyue/docker-ce.repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo && sudo mv /home/fengyue/docker-ce.repo /etc/yum.repos.d/docker-ce.repo " -v

ansible docker_nodes -m shell -a "sudo yum install -y docker-ce" -v
ansible docker_nodes -m shell -a "sudo systemctl start docker" -v
ansible docker_nodes -m shell -a "sudo systemctl enable docker" -v

ansible docker_nodes -m shell -a "docker --version" -v


ansible docker_nodes -m shell -a " curl -L "https://github.com/docker/compose/releases/download/2.27.1/docker-compose-$(uname -s)-$(uname -m)" -o /home/fengyue/docker-compose " -v

sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

ansible docker_nodes -m shell -a " sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose " -v


## nginx



## 
ansible docker_nodes -m copy -a " src=consul.nginx.conf dest=/home/fengyue/deploy/nginx/ " -v
ansible docker_nodes -m copy -a " src=nginx/appsvr_ws_gw.conf dest=/home/fengyue/deploy/nginx/ " -v
ansible docker_nodes -m copy -a " src=nginx/iot_register.conf dest=/home/fengyue/deploy/nginx/ " -v



ansible docker_nodes -m fetch -a "src=/home/fengyue/deploy/key dest=./consul_key flat=yes" -v

## 日志.

ansible docker_nodes -m shell -a " sudo docker logs consul-client-2 -f -n100 " -v

ansible docker_nodes -m shell -a "  " -v



## 创建用户
sudo useradd -m -s /bin/bash fengyue
sudo passwd fengyue
sudo usermod -aG wheel fengyue

sudo visudo => fengyue  ALL=(ALL) NOPASSWD: ALL
su - fengyue

ssh-keygen -t rsa -b 2048
ssh-copy-id fengyue@139.162.61.214


## 防火墙
sudo systemctl status firewalld
sudo firewall-cmd --list-all

sudo firewall-cmd --zone=public --add-port=48080/tcite --permanent

sudo firewall-cmd --reload

sudo firewall-cmd --zone=public --remove-port=80/tcp --permanent


sudo vim /etc/nginx/nginx.conf
sudo systemctl restart nginx
