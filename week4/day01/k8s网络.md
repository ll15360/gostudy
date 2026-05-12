# 这个文件主要是对k8s网络的实践的记录，包括这个镜像的构建，deployment，service的实现
如何通过k8s部署一个微服务,假设当前有一个简单的微服务系统，包括一个网关服务，nacos配置注册中心，服务A,服务B，服务C，和一个聚合服务


一、核心前置：K8s Pod 网络基础
同一个 Pod 内的多个容器：共享同一个 Pod IP，通过不同端口通信；
集群内不同 Pod：IP 属于同一集群网段，原生可 IP + 端口互通，但Pod IP 会随重建、调度漂移，极不稳定，不能直接用于微服务长期通信；
所以必须依赖 Service 提供固定访问入口 + 负载均衡 + DNS 域名解析。


二、三种 Service 类型原理 & 微服务适配

# 方式 1：ClusterIP（集群内部默认类型）
原理Service 通过 selector 匹配标签相同的一组 Pod，分配一个集群内部虚拟固定 IP（ClusterIP）；K8s 内置 CoreDNS 自动做域名解析，集群内可直接通过 Service 名称 + 端口 访问，无需记 Pod IP、无需关心 Pod 漂移。
适用场景只允许集群内部微服务互相调用，不对外网暴露。适配组件：Nacos 注册配置中心、服务 A、服务 B、服务 C、聚合服务。
调用逻辑所有微服务不用写 Nacos 的 Pod IP，直接配置：nacos-service:8848 即可完成注册、配置拉取；内部服务间调用也通过各自 Service 名 + 端口访问。



# 方式 2：NodePort（测试环境外网暴露）
原理NodePort 内置自带 ClusterIP；在集群所有节点上开放一个固定端口（默认区间 30000~32767），外网通过任意节点 IP + NodePort 端口访问；流量链路：外网 → 节点 IP:NodePort → 转发到集群 ClusterIP → 负载均衡到后端 Pod。
适用场景测试环境暴露网关服务，作为外网进入微服务集群的唯一入口。适配组件：网关服务（测试环境首选）。
缺点端口固定区间限制、依赖节点 IP，生产环境不推荐。


# 方式 3：LoadBalancer（生产环境外网暴露）
原理基于 NodePort 实现，云厂商专属；云服务商提供一个独立公网 LB IP，公网访问 LB 公网 IP；流量链路：公网 LB IP → 集群节点 NodePort → 转发到 ClusterIP → 负载均衡到后端业务 Pod。


# 3 镜像的构建: Dockerfile



# 创建POD与数据卷挂载

