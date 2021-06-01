# CloudLab
## Overview
The basic idea is to give a web portal for develper  which can support create VM, k8s, some middle softwares in an very fast and simple way

## Architecture Diagram
![architecture diagram](./cloudLab.png)
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
