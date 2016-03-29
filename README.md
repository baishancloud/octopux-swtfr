swtfr介绍
=========

我们将falcon－transfer用于流量采集系统中
influxdb的写入接口。如果关注openfalcon项目，可以发现最新版本的已经有opentsdb接口。其实我们也实验了opentsdb数据来存储流量信息。但在使用中，我们发现opentsdb单机部署的情况下查询效率有点慢，所以我们实现了influxdb作为tsdb数据库。这个选择有优点也有缺点。列举一下，以便大家选择。请所以慎重选择这个版本。

influxdb：

优点：

1.  单机部署简单。

    1.  查询效率优于opentsdb。（只测试了单机，查询时间快很多，具体数值缺失）

        1.  自带capacitor，容易做一些报警分析。

        2.  备份容易。

    缺点：

    1.  版本比较新，有些功能未完善，还未到1.0版本。

        1.  下一个开源版本不再支持集群。

        2.  查询函数支持比较少。

        3.  版本更新很快，居然换了底层存储组件。备份有旧数据处理问题。

 

opentsdb

优点：

1.  版本较成熟。

    1.  支持函数多。

    缺点：

    1.  依赖hbase，部署比较复杂。

        1.  查询较慢。单机调优效果不是很明显。

 

\----------

Introduction
------------

数据收集，是监控系统一个最基本的功能，在Open-Falcon中，Agent采集到的数据，会先发送给Transfer组件。Transfer在接收到客户端发送的数据，做一些数据规整，检查之后，转发到多个后端系统去处理。在转发到每个后端业务系统的时候，Transfer会根据一致性哈希算法，进行数据分片，来达到后端业务系统的水平扩展。Transfer自身是无状态的，挂掉一台或者多台不会有任何影响。

Transfer支持的业务后端，有三种，Judge、Graph、OpenTSDB(开源版本尚未开放此功能)。Judge是我们开发的高性能告警判定组件，Graph是我们开发的高性能数据存储、归档、查询组件，OpenTSDB是开源的时间序列数据存储服务。每个业务后端，都可以通过Transfer的配置文件来开启。

Transfer的数据来源，一般有四种：

1.Falcon-agent主动采集的基础监控数据。
2.Falcon-agent执行用户自定义的插件返回的数据。
3.client-library：线上的业务系统，都嵌入使用了统一的基础库，对于业务系统中每个业务接口，都会主动计算其qps、latency等指标，并上报。
4.用户产生的一些自定义的指标，由用户自行上报。

这四种数据，都会先发送给本机的Proxy-gateway，再由Proxy-gateway转发给Transfer

一个推送数据给Proxy-gateway的例子：

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ python
#!-*- coding:utf8 -*-
    
import requests
import time
import json

ts = int(time.time())
payload = [
    {
        "endpoint": "test-endpoint",
        "metric": "test-metric",
        "timestamp": ts,
        "step": 60,
        "value": 1,
        "counterType": "GAUGE",
        "tags": "location=beijing,service=falcon",
    },
]
r=requests.post("http://127.0.0.1:1988/v1/push",data=json.dumps(payload))
print r.text
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Installation
------------

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ bash
# set $GOPATH and $GOROOT

mkdir -p $GOPATH/src/github.com/open-falcon
cd $GOPATH/src/github.com/open-falcon
git clone https://github.com/open-falcon/transfer.git

cd transfer
go get ./...
./control build

./control start
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Usage
-----

send items via transfer's http-api

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ bash
#!/bin/bash
e="test.endpoint.1" 
m="test.metric.1"
t="t0=tag0,t1=tag1,t2=tag2"
ts=`date +%s`
curl -s -X POST -d "[{\"metric\":\"$m\", \"endpoint\":\"$e\", \"timestamp\":$ts,\"step\":60, \"value\":9, \"counterType\":\"GAUGE\",\"tags\":\"$t\"}]" "127.0.0.1:6060/api/push" | python -m json.tool
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

u want sending items via python jsonrpc client? turn to one python example:
`./test/rcpclient.py`

u want sending items via java jsonrpc client? turn to one java example:
[jsonrpc4go](<https://github.com/niean/jsonrpc4go>)

Configuration
-------------

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
debug: true/false, 如果为true，日志中会打印debug信息

http
    - enable: true/false, 表示是否开启该http端口，该端口为控制端口，主要用来对transfer发送控制命令、统计命令、debug命令等
    - listen: 表示监听的http端口

rpc
    - enable: true/false, 表示是否开启该jsonrpc数据接收端口, Agent发送数据使用的就是该端口
    - listen: 表示监听的http端口

socket #即将被废弃,请避免使用
    - enable: true/false, 表示是否开启该telnet方式的数据接收端口，这是为了方便用户一行行的发送数据给transfer
    - listen: 表示监听的http端口

judge
    - enable: true/false, 表示是否开启向judge发送数据
    - batch: 数据转发的批量大小，可以加快发送速度，建议保持默认值
    - connTimeout: 单位是毫秒，与后端建立连接的超时时间，可以根据网络质量微调，建议保持默认
    - callTimeout: 单位是毫秒，发送数据给后端的超时时间，可以根据网络质量微调，建议保持默认
    - pingMethod: 后端提供的ping接口，用来探测连接是否可用，必须保持默认
    - maxConns: 连接池相关配置，最大连接数，建议保持默认
    - maxIdle: 连接池相关配置，最大空闲连接数，建议保持默认
    - replicas: 这是一致性hash算法需要的节点副本数量，建议不要变更，保持默认即可
    - cluster: key-value形式的字典，表示后端的judge列表，其中key代表后端judge名字，value代表的是具体的ip:port

graph
    - enable: true/false, 表示是否开启向graph发送数据
    - batch: 数据转发的批量大小，可以加快发送速度，建议保持默认值
    - connTimeout: 单位是毫秒，与后端建立连接的超时时间，可以根据网络质量微调，建议保持默认
    - callTimeout: 单位是毫秒，发送数据给后端的超时时间，可以根据网络质量微调，建议保持默认
    - pingMethod: 后端提供的ping接口，用来探测连接是否可用，必须保持默认
    - maxConns: 连接池相关配置，最大连接数，建议保持默认
    - maxIdle: 连接池相关配置，最大空闲连接数，建议保持默认
    - replicas: 这是一致性hash算法需要的节点副本数量，建议不要变更，保持默认即可
    - migrating: true/false，当我们需要对graph后端列表进行扩容的时候，设置为true, transfer会根据扩容前后的实例信息，对每个数据采集项，进行两次一致性哈希计算，根据计算结果，来决定是否需要发送双份的数据，当新扩容的服务器积累了足够久的数据后，就可以设置为false。
    - cluster: key-value形式的字典，表示后端的graph列表，其中key代表后端graph名字，value代表的是具体的ip:port(多个地址 用逗号隔开, transfer会将同一份数据 发送至各个地址)
    - clusterMigrating: key-value形式的字典，表示新扩容的后端的graph列表，其中key代表后端graph名字，value代表的是具体的ip:port(多个地址 用逗号隔开, transfer会将同一份数据 发送至各个地址)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
