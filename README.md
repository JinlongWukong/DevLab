# CloudLab
## Overview
The basic idea is to give a web portal for develper  which can support create VM, k8s, some middle softwares in an very fast and simple way

## Architecture Diagram
![architecture diagram](./cloudLab.png)

## Features
- Virtual Machine management(libvirt, kvm)
- External Access(iptables dnat)
- Auto Lifecycle management
- In-Memory Persistant
- Webex events Notification
- K8s cluster management(Todo)
- Middleware management(Todo)

## Installation
- controller 
#### Build docker image
```
docker build -t controller .
```
#### Run container
```
docker run -d --net host --rm --env HTTPS_PROXY=xxxxx --env NO_PROXY="xxxx" --env BOT_TOKEN=xxxxx -v "$(pwd)"/node.json:/app/node.json -v "$(pwd)"/account.json:/app/account.json -v "$(pwd)"/config.ini:/app/config.ini controller
```
- deployer


How to install deployer? refer to [Deployer repo](https://github.com/JinlongWukong/CloudLab-ansible)

## How to use
- Edit config.ini 
- Run docker container
- Open web brower -> http://<cloudLab ip>:8088/