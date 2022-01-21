---
home: true
heroImage: "https://user-images.githubusercontent.com/8912557/139633211-96844d27-55d7-44a9-9cdc-5aea96441613.png"
tagline: 把整个集群看成一台服务器，把kubernetes看成云操作系统，吸取docker设计精髓实现分布式软件镜像化构建、交付、运行
actionText: 快速开始→
actionLink: "/zh/getting-started/introduction"
features:

- title: 构建
  details: 使用Kubefile定义整个集群所有依赖，把kubernetes 中间件 数据库和SaaS软件所有依赖打包到集群镜像中
- title: 交付
  details: 像交付Docker镜像一样交付整个集群和集群中的分布式软件
- title: 运行
  details: 一条命令启动整个集群, 集群纬度保障一致性，兼容主流linux系统与多种架构
footer: "钉钉讨论组: 34619594 微信: fangnux"
---

# 构建&运行一个自定义kuberentes集群

本例中介绍如何构建一个包含dashboard的集群镜像，然后一键启动。

[![asciicast](https://asciinema.org/a/446106.svg)](https://asciinema.org/a/446106?speed=3)