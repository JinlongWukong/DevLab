# DevLab
## Overview
This is a IPSaaS platform for develpers, Its basic idea is to give a web portal to develper which can support create VM, k8s, common middleware in a faster and simpler way

## Architecture Diagram
![architecture diagram](./views/image/DevLab.png)

## Features
- Virtual Machine Management(libvirt, kvm)
- Multi-Node Inter-connection(hostgw)
- External Access(iptables dnat)
- Auto vm Lifecycle Management
- In-Memory Persistant
- Webex/Telegram Events Notification
- K8s Cluster Management
- SaaS Management
- Account Management
- Token Authentication
- Web Terminal(ssh, novnc)
- Node scheduler algorithm(random, weight)

## Installation
- controller 
#### Build docker image
```
docker build -t controller --build-arg https_proxy=xxxxx .
```
#### Run container
```
#Download example config.ini from github
vim config.ini
mkdir .db/
docker run -d --name devlab_controller --net host --env HTTPS_PROXY=xxxxx --env NO_PROXY="xxxx" --env BOT_TOKEN=xxxxx -v "$(pwd)"/.db/:/app/.db -v "$(pwd)"/config.ini:/app/config.ini controller
```
- deployer

How to install deployer? refer to [Deployer repo](https://github.com/JinlongWukong/DevLab-ansible)

- novnc

This is optional component
```
docker run -d --net host geek1011/easy-novnc  -H -P
```

## How to use
- Edit config.ini 
- Run docker container
- Open web brower -> http://ip:8088/
