# kubernetes 官方教程

**[Kubernetes](https://kubernetes.io/zh-cn/docs/concepts/overview/) 也称为 K8s，是用于自动部署、扩缩和管理容器化应用程序的开源系统。**

> Kubernetes 是一个生产级别容器编排系统，可协调在计算机集群内和跨计算机集群的应用容器的部署（调度）和执行。
>
> 它将组成应用程序的容器组合成逻辑单元，以便于管理和服务发现。Kubernetes 源自[Google 15 年生产环境的运维经验](http://queue.acm.org/detail.cfm?id=2898444)，同时凝聚了社区的最佳创意和实践。

- 星际尺度：Google 每周运行数十亿个容器，Kubernetes 基于与之相同的原则来设计，能够在不扩张运维团队的情况下进行规模扩展。
- 处处适用：无论是本地测试，还是跨国公司，Kubernetes 的灵活性都能让你在应对复杂系统时得心应手。
- 永不过时：Kubernetes 是开源系统，可以自由地部署在企业内部，私有云、混合云或公有云，让您轻松地做出合适的选择。

## [概述](https://kubernetes.io/zh-cn/docs/concepts/overview/)

**Kubernetes 是一个可移植、可扩展的开源平台，用于管理容器化的工作负载和服务，方便进行声明式配置和自动化。**Kubernetes 拥有一个庞大且快速增长的生态系统，其服务、支持和工具的使用范围广泛。

### Kubernetes 组件

当你部署完 Kubernetes，便拥有了一个完整的集群。

一组工作机器，称为 [节点](https://kubernetes.io/zh-cn/docs/concepts/architecture/nodes/)， 会运行容器化应用程序。每个集群至少有一个工作节点。

工作节点会托管 [Pod](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/) ，而 Pod 就是作为应用负载的组件。 [控制平面](https://kubernetes.io/zh-cn/docs/reference/glossary/?all=true#term-control-plane)管理集群中的工作节点和 Pod。 在生产环境中，控制平面通常跨多台计算机运行， 一个集群通常运行多个节点，提供容错性和高可用性。

本文档概述了一个正常运行的 Kubernetes 集群所需的各种组件。

![Kubernetes 的组件](images/components-of-kubernetes.svg)

Kubernetes 集群的组件

#### 控制平面组件（Control Plane Components）

控制平面组件会为集群做出全局决策，比如资源的调度。 以及检测和响应集群事件，例如当不满足部署的 `replicas` 字段时， 要启动新的 [pod](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/)）。

控制平面组件可以在集群中的任何节点上运行。 然而，为了简单起见，设置脚本通常会在同一个计算机上启动所有控制平面组件， 并且不会在此计算机上运行用户容器。 请参阅[使用 kubeadm 构建高可用性集群](https://kubernetes.io/zh-cn/docs/setup/production-environment/tools/kubeadm/high-availability/) 中关于跨多机器控制平面设置的示例。

#### kube-apiserver

API 服务器是 Kubernetes [控制平面](https://kubernetes.io/zh-cn/docs/reference/glossary/?all=true#term-control-plane)的组件， 该组件负责公开了 Kubernetes API，负责处理接受请求的工作。 API 服务器是 Kubernetes 控制平面的前端。

Kubernetes API 服务器的主要实现是 [kube-apiserver](https://kubernetes.io/zh-cn/docs/reference/command-line-tools-reference/kube-apiserver/)。 `kube-apiserver` 设计上考虑了水平扩缩，也就是说，它可通过部署多个实例来进行扩缩。 你可以运行 `kube-apiserver` 的多个实例，并在这些实例之间平衡流量。

#### etcd

一致且高度可用的键值存储，用作 Kubernetes 的所有集群数据的后台数据库。

如果你的 Kubernetes 集群使用 etcd 作为其后台数据库， 请确保你针对这些数据有一份 [备份](https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/configure-upgrade-etcd/#backing-up-an-etcd-cluster)计划。

你可以在官方[文档](https://etcd.io/docs/)中找到有关 etcd 的深入知识。

#### kube-scheduler

`kube-scheduler` 是[控制平面](https://kubernetes.io/zh-cn/docs/reference/glossary/?all=true#term-control-plane)的组件， 负责监视新创建的、未指定运行[节点（node）](https://kubernetes.io/zh-cn/docs/concepts/architecture/nodes/)的 [Pods](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/)， 并选择节点来让 Pod 在上面运行。

调度决策考虑的因素包括单个 Pod 及 Pods 集合的资源需求、软硬件及策略约束、 亲和性及反亲和性规范、数据位置、工作负载间的干扰及最后时限。

#### kube-controller-manager

[kube-controller-manager](https://kubernetes.io/zh-cn/docs/reference/command-line-tools-reference/kube-controller-manager/) 是[控制平面](https://kubernetes.io/zh-cn/docs/reference/glossary/?all=true#term-control-plane)的组件， 负责运行[控制器](https://kubernetes.io/zh-cn/docs/concepts/architecture/controller/)进程。

从逻辑上讲， 每个[控制器](https://kubernetes.io/zh-cn/docs/concepts/architecture/controller/)都是一个单独的进程， 但是为了降低复杂性，它们都被编译到同一个可执行文件，并在同一个进程中运行。

这些控制器包括：

- 节点控制器（Node Controller）：负责在节点出现故障时进行通知和响应
- 任务控制器（Job Controller）：监测代表一次性任务的 Job 对象，然后创建 Pods 来运行这些任务直至完成
- 端点分片控制器（EndpointSlice controller）：填充端点分片（EndpointSlice）对象（以提供 Service 和 Pod 之间的链接）。
- 服务账号控制器（ServiceAccount controller）：为新的命名空间创建默认的服务账号（ServiceAccount）。

#### cloud-controller-manager

一个 Kubernetes [控制平面](https://kubernetes.io/zh-cn/docs/reference/glossary/?all=true#term-control-plane)组件， 嵌入了特定于云平台的控制逻辑。 云控制器管理器（Cloud Controller Manager）允许你将你的集群连接到云提供商的 API 之上， 并将与该云平台交互的组件同与你的集群交互的组件分离开来。

`cloud-controller-manager` 仅运行特定于云平台的控制器。 因此如果你在自己的环境中运行 Kubernetes，或者在本地计算机中运行学习环境， 所部署的集群不需要有云控制器管理器。

与 `kube-controller-manager` 类似，`cloud-controller-manager` 将若干逻辑上独立的控制回路组合到同一个可执行文件中， 供你以同一进程的方式运行。 你可以对其执行水平扩容（运行不止一个副本）以提升性能或者增强容错能力。

下面的控制器都包含对云平台驱动的依赖：

- 节点控制器（Node Controller）：用于在节点终止响应后检查云提供商以确定节点是否已被删除
- 路由控制器（Route Controller）：用于在底层云基础架构中设置路由
- 服务控制器（Service Controller）：用于创建、更新和删除云提供商负载均衡器

#### Node 组件

节点组件会在每个节点上运行，负责维护运行的 Pod 并提供 Kubernetes 运行环境。

#### kubelet

`kubelet` 会在集群中每个[节点（node）](https://kubernetes.io/zh-cn/docs/concepts/architecture/nodes/)上运行。 它保证[容器（containers）](https://kubernetes.io/zh-cn/docs/concepts/overview/what-is-kubernetes/#why-containers)都运行在 [Pod](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/) 中。

kubelet 接收一组通过各类机制提供给它的 PodSpecs， 确保这些 PodSpecs 中描述的容器处于运行状态且健康。 kubelet 不会管理不是由 Kubernetes 创建的容器。

#### kube-proxy

[kube-proxy](https://kubernetes.io/zh-cn/docs/reference/command-line-tools-reference/kube-proxy/) 是集群中每个[节点（node）](https://kubernetes.io/zh-cn/docs/concepts/architecture/nodes/)上所运行的网络代理， 实现 Kubernetes [服务（Service）](https://kubernetes.io/zh-cn/docs/concepts/services-networking/service/) 概念的一部分。

kube-proxy 维护节点上的一些网络规则， 这些网络规则会允许从集群内部或外部的网络会话与 Pod 进行网络通信。

如果操作系统提供了可用的数据包过滤层，则 kube-proxy 会通过它来实现网络规则。 否则，kube-proxy 仅做流量转发。

#### 容器运行时（Container Runtime）

容器运行环境是负责运行容器的软件。

Kubernetes 支持许多容器运行环境，例如 [containerd](https://containerd.io/docs/)、 [CRI-O](https://cri-o.io/#what-is-cri-o) 以及 [Kubernetes CRI (容器运行环境接口)](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-node/container-runtime-interface.md) 的其他任何实现。

#### 插件（Addons）

插件使用 Kubernetes 资源（[DaemonSet](https://kubernetes.io/zh-cn/docs/concepts/workloads/controllers/daemonset/)、 [Deployment](https://kubernetes.io/zh-cn/docs/concepts/workloads/controllers/deployment/) 等）实现集群功能。 因为这些插件提供集群级别的功能，插件中命名空间域的资源属于 `kube-system` 命名空间。

下面描述众多插件中的几种。有关可用插件的完整列表，请参见 [插件（Addons）](https://kubernetes.io/zh-cn/docs/concepts/cluster-administration/addons/)。

#### DNS

尽管其他插件都并非严格意义上的必需组件，但几乎所有 Kubernetes 集群都应该有[集群 DNS](https://kubernetes.io/zh-cn/docs/concepts/services-networking/dns-pod-service/)， 因为很多示例都需要 DNS 服务。

集群 DNS 是一个 DNS 服务器，和环境中的其他 DNS 服务器一起工作，它为 Kubernetes 服务提供 DNS 记录。

Kubernetes 启动的容器自动将此 DNS 服务器包含在其 DNS 搜索列表中。

#### Web 界面（仪表盘）

[Dashboard](https://kubernetes.io/zh-cn/docs/tasks/access-application-cluster/web-ui-dashboard/) 是 Kubernetes 集群的通用的、基于 Web 的用户界面。 它使用户可以管理集群中运行的应用程序以及集群本身， 并进行故障排除。

#### 容器资源监控

[容器资源监控](https://kubernetes.io/zh-cn/docs/tasks/debug/debug-cluster/resource-usage-monitoring/) 将关于容器的一些常见的时间序列度量值保存到一个集中的数据库中， 并提供浏览这些数据的界面。

#### 集群层面日志

[集群层面日志](https://kubernetes.io/zh-cn/docs/concepts/cluster-administration/logging/)机制负责将容器的日志数据保存到一个集中的日志存储中， 这种集中日志存储提供搜索和浏览接口。

### Kubernetes API

Kubernetes [控制面](https://kubernetes.io/zh-cn/docs/reference/glossary/?all=true#term-control-plane)的核心是 [API 服务器](https://kubernetes.io/zh-cn/docs/concepts/overview/components/#kube-apiserver)。 API 服务器负责提供 HTTP API，以供用户、集群中的不同部分和集群外部组件相互通信。

Kubernetes API 使你可以查询和操纵 Kubernetes API 中对象（例如：Pod、Namespace、ConfigMap 和 Event）的状态。

大部分操作都可以通过 [kubectl](https://kubernetes.io/zh-cn/docs/reference/kubectl/) 命令行接口或类似 [kubeadm](https://kubernetes.io/zh-cn/docs/reference/setup-tools/kubeadm/) 这类命令行工具来执行， 这些工具在背后也是调用 API。不过，你也可以使用 REST 调用来访问这些 API。

如果你正在编写程序来访问 Kubernetes API， 可以考虑使用[客户端库](https://kubernetes.io/zh-cn/docs/reference/using-api/client-libraries/)之一。

#### OpenAPI 规范

完整的 API 细节是用 [OpenAPI](https://www.openapis.org/) 来表述的。

#### OpenAPI V2

Kubernetes API 服务器通过 `/openapi/v2` 端点提供聚合的 OpenAPI v2 规范。 你可以按照下表所给的请求头部，指定响应的格式：

| 头部               | 可选值                                                       | 说明                     |
| ------------------ | ------------------------------------------------------------ | ------------------------ |
| `Accept-Encoding`  | `gzip`                                                       | *不指定此头部也是可以的* |
| `Accept`           | `application/com.github.proto-openapi.spec.v2@v1.0+protobuf` | *主要用于集群内部*       |
| `application/json` | *默认值*                                                     |                          |
| `*`                | *提供*`application/json`                                     |                          |

Kubernetes 为 API 实现了一种基于 Protobuf 的序列化格式，主要用于集群内部通信。 关于此格式的详细信息，可参考 [Kubernetes Protobuf 序列化](https://git.k8s.io/design-proposals-archive/api-machinery/protobuf.md)设计提案。 每种模式对应的接口描述语言（IDL）位于定义 API 对象的 Go 包中。

#### OpenAPI V3

**特性状态：** `Kubernetes v1.24 [beta]`

Kubernetes v1.25 提供将其 API 以 OpenAPI v3 形式发布的 beta 支持； 这一功能特性处于 beta 状态，默认被开启。 你可以通过为 kube-apiserver 组件关闭 `OpenAPIV3` [特性门控](https://kubernetes.io/zh-cn/docs/reference/command-line-tools-reference/feature-gates/)来禁用此 beta 特性。

发现端点 `/openapi/v3` 被提供用来查看可用的所有组、版本列表。 此列表仅返回 JSON。这些组、版本以下面的格式提供：

```yaml
{
    "paths": {
        ...,
        "api/v1": {
            "serverRelativeURL": "/openapi/v3/api/v1?hash=CC0E9BFD992D8C59AEC98A1E2336F899E8318D3CF4C68944C3DEC640AF5AB52D864AC50DAA8D145B3494F75FA3CFF939FCBDDA431DAD3CA79738B297795818CF"
        },
        "apis/admissionregistration.k8s.io/v1": {
            "serverRelativeURL": "/openapi/v3/apis/admissionregistration.k8s.io/v1?hash=E19CC93A116982CE5422FC42B590A8AFAD92CDE9AE4D59B5CAAD568F083AD07946E6CB5817531680BCE6E215C16973CD39003B0425F3477CFD854E89A9DB6597"
        },
        ....
    }
}
```

为了改进客户端缓存，相对的 URL 会指向不可变的 OpenAPI 描述。 为了此目的，API 服务器也会设置正确的 HTTP 缓存标头 （`Expires` 为未来 1 年，和 `Cache-Control` 为 `immutable`）。 当一个过时的 URL 被使用时，API 服务器会返回一个指向最新 URL 的重定向。

Kubernetes API 服务器会在端点 `/openapi/v3/apis/<group>/<version>?hash=<hash>` 发布一个 Kubernetes 组版本的 OpenAPI v3 规范。

请参阅下表了解可接受的请求头部。

| 头部               | 可选值                                                       | 说明                       |
| ------------------ | ------------------------------------------------------------ | -------------------------- |
| `Accept-Encoding`  | `gzip`                                                       | *不提供此头部也是可接受的* |
| `Accept`           | `application/com.github.proto-openapi.spec.v3@v1.0+protobuf` | *主要用于集群内部使用*     |
| `application/json` | *默认*                                                       |                            |
| `*`                | *以* `application/json` 形式返回                             |                            |

#### 持久化

Kubernetes 通过将序列化状态的对象写入到 [etcd](https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/configure-upgrade-etcd/) 中完成存储操作。

#### API 组和版本控制

为了更容易消除字段或重组资源的呈现方式，Kubernetes 支持多个 API 版本，每个版本位于不同的 API 路径， 例如 `/api/v1` 或 `/apis/rbac.authorization.k8s.io/v1alpha1`。

版本控制是在 API 级别而不是在资源或字段级别完成的，以确保 API 呈现出清晰、一致的系统资源和行为视图， 并能够控制对生命结束和/或实验性 API 的访问。

为了更容易演进和扩展其 API，Kubernetes 实现了 [API 组](https://kubernetes.io/zh-cn/docs/reference/using-api/#api-groups)， 这些 API 组可以被[启用或禁用](https://kubernetes.io/zh-cn/docs/reference/using-api/#enabling-or-disabling)。

API 资源通过其 API 组、资源类型、名字空间（用于名字空间作用域的资源）和名称来区分。 API 服务器透明地处理 API 版本之间的转换：所有不同的版本实际上都是相同持久化数据的呈现。 API 服务器可以通过多个 API 版本提供相同的底层数据。

例如，假设针对相同的资源有两个 API 版本：`v1` 和 `v1beta1`。 如果你最初使用其 API 的 `v1beta1` 版本创建了一个对象， 你稍后可以使用 `v1beta1` 或 `v1` API 版本来读取、更新或删除该对象， 直到 `v1beta1` 版本被废弃和移除为止。此后，你可以使用 `v1` API 继续访问和修改该对象。

#### API 变更

任何成功的系统都要随着新的使用案例的出现和现有案例的变化来成长和变化。 为此，Kubernetes 已设计了 Kubernetes API 来持续变更和成长。 Kubernetes 项目的目标是 **不要** 给现有客户端带来兼容性问题，并在一定的时期内维持这种兼容性， 以便其他项目有机会作出适应性变更。

一般而言，新的 API 资源和新的资源字段可以被频繁地添加进来。 删除资源或者字段则要遵从 [API 废弃策略](https://kubernetes.io/zh-cn/docs/reference/using-api/deprecation-policy/)。

Kubernetes 对维护达到正式发布（GA）阶段的官方 API 的兼容性有着很强的承诺，通常这一 API 版本为 `v1`。 此外，Kubernetes 保持与 Kubernetes 官方 API 的 **Beta** API 版本持久化数据的兼容性， 并确保在该功能特性已进入稳定期时数据可以通过 GA API 版本进行转换和访问。

如果你采用一个 Beta API 版本，一旦该 API 进阶，你将需要转换到后续的 Beta 或稳定的 API 版本。 执行此操作的最佳时间是 Beta API 处于弃用期，因为此时可以通过两个 API 版本同时访问那些对象。 一旦 Beta API 结束其弃用期并且不再提供服务，则必须使用替换的 API 版本。

**说明：**

尽管 Kubernetes 也努力为 **Alpha** API 版本维护兼容性，在有些场合兼容性是无法做到的。 如果你使用了任何 Alpha API 版本，需要在升级集群时查看 Kubernetes 发布说明， 如果 API 确实以不兼容的方式发生变更，则需要在升级之前删除所有现有的 Alpha 对象。

关于 API 版本分级的定义细节，请参阅 [API 版本参考](https://kubernetes.io/zh-cn/docs/reference/using-api/#api-versioning)页面。

#### API 扩展

有两种途径来扩展 Kubernetes API：

1. 你可以使用[自定义资源](https://kubernetes.io/zh-cn/docs/concepts/extend-kubernetes/api-extension/custom-resources/)来以声明式方式定义 API 服务器如何提供你所选择的资源 API。
2. 你也可以选择实现自己的[聚合层](https://kubernetes.io/zh-cn/docs/concepts/extend-kubernetes/api-extension/apiserver-aggregation/)来扩展 Kubernetes API。

#### 接下来

- 了解如何通过添加你自己的 [CustomResourceDefinition](https://kubernetes.io/zh-cn/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/) 来扩展 Kubernetes API。
- [控制 Kubernetes API 访问](https://kubernetes.io/zh-cn/docs/concepts/security/controlling-access/)页面描述了集群如何针对 API 访问管理身份认证和鉴权。
- 通过阅读 [API 参考](https://kubernetes.io/zh-cn/docs/reference/kubernetes-api/)了解 API 端点、资源类型以及示例。
- 阅读 [API 变更（英文）](https://git.k8s.io/community/contributors/devel/sig-architecture/api_changes.md#readme) 以了解什么是兼容性的变更以及如何变更 API。

## 学习 Kubernetes 基础知识

### Kubernetes 可以为你做些什么?

通过现代的 Web 服务，用户希望应用程序能够 24/7 全天候使用，开发人员希望每天可以多次发布部署新版本的应用程序。 容器化可以帮助软件包达成这些目标，使应用程序能够以简单快速的方式发布和更新，而无需停机。Kubernetes 帮助你确保这些容器化的应用程序在你想要的时间和地点运行，并帮助应用程序找到它们需要的资源和工具。Kubernetes 是一个可用于生产的开源平台，根据 Google 容器集群方面积累的经验，以及来自社区的最佳实践而设计。

## 创建集群

### Kubernetes 集群

**Kubernetes 协调一个高可用计算机集群，每个计算机作为独立单元互相连接工作。** Kubernetes 中的抽象允许你将容器化的应用部署到集群，而无需将它们绑定到某个特定的独立计算机。为了使用这种新的部署模型，应用需要以将应用与单个主机分离的方式打包：它们需要被容器化。与过去的那种应用直接以包的方式深度与主机集成的部署模型相比，容器化应用更灵活、更可用。 **Kubernetes 以更高效的方式跨集群自动分发和调度应用容器。** Kubernetes 是一个开源平台，并且可应用于生产环境。

一个 Kubernetes 集群包含两种类型的资源:

- **Master** 调度整个集群
- **Nodes** 负责运行应用

#### 集群图

![img](images/module_01_cluster.svg)



**Master 负责管理整个集群。** Master 协调集群中的所有活动，例如调度应用、维护应用的所需状态、应用扩容以及推出新的更新。

**Node 是一个虚拟机或者物理机，它在 Kubernetes 集群中充当工作机器的角色** 每个Node都有 Kubelet , 它管理 Node 而且是 Node 与 Master 通信的代理。 Node 还应该具有用于处理容器操作的工具，例如 Docker 或 rkt 。处理生产级流量的 Kubernetes 集群至少应具有三个 Node，因为如果一个 Node 出现故障其对应的 etcd 成员和控制平面实例都会丢失，并且冗余会受到影响。 你可以通过添加更多控制平面节点来降低这种风险 。

*Master 管理集群，Node 用于托管正在运行的应用。*

在 Kubernetes 上部署应用时，你告诉 Master 启动应用容器。 Master 就编排容器在集群的 Node 上运行。 **Node 使用 Master 暴露的 Kubernetes API 与 Master 通信。**终端用户也可以使用 Kubernetes API 与集群交互。

Kubernetes 既可以部署在物理机上也可以部署在虚拟机上。你可以使用 Minikube 开始部署 Kubernetes 集群。 Minikube 是一种轻量级的 Kubernetes 实现，可在本地计算机上创建 VM 并部署仅包含一个节点的简单集群。 Minikube 可用于 Linux ， macOS 和 Windows 系统。Minikube CLI 提供了用于引导集群工作的多种操作，包括启动、停止、查看状态和删除。在本教程里，你可以使用预装有 Minikube 的在线终端进行体验。

> Master 管理集群，Node 用于托管正在运行的应用。

### 部署应用

### Kubernetes 部署

一旦运行了 Kubernetes 集群，就可以在其上部署容器化应用程序。 为此，你需要创建 Kubernetes **Deployment** 配置。Deployment 指挥 Kubernetes 如何创建和更新应用程序的实例。创建 Deployment 后，Kubernetes master 将应用程序实例调度到集群中的各个节点上。

创建应用程序实例后，Kubernetes Deployment 控制器会持续监视这些实例。 如果托管实例的节点关闭或被删除，则 Deployment 控制器会将该实例替换为集群中另一个节点上的实例。 **这提供了一种自我修复机制来解决机器故障维护问题。**

在没有 Kubernetes 这种编排系统之前，安装脚本通常用于启动应用程序，但它们不允许从机器故障中恢复。通过创建应用程序实例并使它们在节点之间运行， Kubernetes Deployments 提供了一种与众不同的应用程序管理方法。

#### 部署你在 Kubernetes 上的第一个应用程序

![img](images/module_02_first_app.svg)



你可以使用 Kubernetes 命令行界面 **Kubectl** 创建和管理 Deployment。Kubectl 使用 Kubernetes API 与集群进行交互。在本单元中，你将学习创建在 Kubernetes 集群上运行应用程序的 Deployment 所需的最常见的 Kubectl 命令。

创建 Deployment 时，你需要指定应用程序的容器镜像以及要运行的副本数。你可以稍后通过更新 Deployment 来更改该信息; 模块 [5](https://kubernetes.io/zh-cn/docs/tutorials/kubernetes-basics/scale-intro/) 和 [6](https://kubernetes.io/zh-cn/docs/tutorials/kubernetes-basics/update-intro/) 讨论了如何扩展和更新 Deployments。

*应用程序需要打包成一种受支持的容器格式，以便部署在 Kubernetes 上*

对于我们的第一次部署，我们将使用打包在 Docker 容器中的 Node.js 应用程序。 要创建 Node.js 应用程序并部署 Docker 容器，请按照 [你好 Minikube 教程](https://kubernetes.io/zh-cn/docs/tutorials/hello-minikube/).

现在你已经了解了 Deployment 的内容，让我们转到在线教程并部署我们的第一个应用程序！

总结：

- *Deployment 负责创建和更新应用程序的实例*
- *应用程序需要打包成一种受支持的容器格式，以便部署在 Kubernetes 上*



### Kubernetes Pods

在模块 [2](https://kubernetes.io/zh-cn/docs/tutorials/kubernetes-basics/deploy-app/deploy-intro/)创建 Deployment 时, Kubernetes 添加了一个 **Pod** 来托管你的应用实例。Pod 是 Kubernetes 抽象出来的，表示一组一个或多个应用程序容器（如 Docker），以及这些容器的一些共享资源。这些资源包括:

- 共享存储，当作卷
- 网络，作为唯一的集群 IP 地址
- 有关每个容器如何运行的信息，例如容器镜像版本或要使用的特定端口。

Pod 为特定于应用程序的“逻辑主机”建模，并且可以包含相对紧耦合的不同应用容器。例如，Pod 可能既包含带有 Node.js 应用的容器，也包含另一个不同的容器，用于提供 Node.js 网络服务器要发布的数据。Pod 中的容器共享 IP 地址和端口，始终位于同一位置并且共同调度，并在同一工作节点上的共享上下文中运行。

Pod是 Kubernetes 平台上的原子单元。 当我们在 Kubernetes 上创建 Deployment 时，该 Deployment 会在其中创建包含容器的 Pod （而不是直接创建容器）。每个 Pod 都与调度它的工作节点绑定，并保持在那里直到终止（根据重启策略）或删除。 如果工作节点发生故障，则会在集群中的其他可用工作节点上调度相同的 Pod。

#### Pod 概览

![img](images/module_03_pods.svg)



#### 工作节点

一个 pod 总是运行在 **工作节点**。工作节点是 Kubernetes 中的参与计算的机器，可以是虚拟机或物理计算机，具体取决于集群。每个工作节点由主节点管理。工作节点可以有多个 pod ，Kubernetes 主节点会自动处理在集群中的工作节点上调度 pod 。 主节点的自动调度考量了每个工作节点上的可用资源。

每个 Kubernetes 工作节点至少运行:

- Kubelet，负责 Kubernetes 主节点和工作节点之间通信的过程; 它管理 Pod 和机器上运行的容器。
- 容器运行时（如 Docker）负责从仓库中提取容器镜像，解压缩容器以及运行应用程序。



#### 工作节点概览

![img](images/module_03_nodes.svg)



#### 使用 kubectl 进行故障排除

在模块 [2](https://kubernetes.io/zh-cn/docs/tutorials/kubernetes-basics/deploy-app/deploy-intro/),你使用了 Kubectl 命令行界面。 你将继续在第3单元中使用它来获取有关已部署的应用程序及其环境的信息。 最常见的操作可以使用以下 kubectl 命令完成：

- **kubectl get** - 列出资源
- **kubectl describe** - 显示有关资源的详细信息
- **kubectl logs** - 打印 pod 和其中容器的日志
- **kubectl exec** - 在 pod 中的容器上执行命令

你可以使用这些命令查看应用程序的部署时间，当前状态，运行位置以及配置。

现在我们了解了有关集群组件和命令行的更多信息，让我们来探索一下我们的应用程序。

总结：

- *如果它们紧耦合并且需要共享磁盘等资源，这些容器应在一个 Pod 中编排。*



### Kubernetes Service 总览

Kubernetes [Pod](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/) 是转瞬即逝的。 Pod 实际上拥有 [生命周期](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/pod-lifecycle/)。 当一个工作 Node 挂掉后, 在 Node 上运行的 Pod 也会消亡。 [ReplicaSet](https://kubernetes.io/zh-cn/docs/concepts/workloads/controllers/replicaset/) 会自动地通过创建新的 Pod 驱动集群回到目标状态，以保证应用程序正常运行。 换一个例子，考虑一个具有3个副本数的用作图像处理的后端程序。这些副本是可替换的; 前端系统不应该关心后端副本，即使 Pod 丢失或重新创建。也就是说，Kubernetes 集群中的每个 Pod (即使是在同一个 Node 上的 Pod )都有一个唯一的 IP 地址，因此需要一种方法自动协调 Pod 之间的变更，以便应用程序保持运行。

Kubernetes 中的服务(Service)是一种抽象概念，它定义了 Pod 的逻辑集和访问 Pod 的协议。Service 使从属 Pod 之间的松耦合成为可能。 和其他 Kubernetes 对象一样, Service 用 YAML [(更推荐)](https://kubernetes.io/zh-cn/docs/concepts/configuration/overview/#general-configuration-tips) 或者 JSON 来定义. Service 下的一组 Pod 通常由 *LabelSelector* (请参阅下面的说明为什么你可能想要一个 spec 中不包含`selector`的服务)来标记。

尽管每个 Pod 都有一个唯一的 IP 地址，但是如果没有 Service ，这些 IP 不会暴露在集群外部。Service 允许你的应用程序接收流量。Service 也可以用在 ServiceSpec 标记`type`的方式暴露

- *ClusterIP* (默认) - 在集群的内部 IP 上公开 Service 。这种类型使得 Service 只能从集群内访问。
- *NodePort* - 使用 NAT 在集群中每个选定 Node 的相同端口上公开 Service 。使用`<NodeIP>:<NodePort>` 从集群外部访问 Service。是 ClusterIP 的超集。
- *LoadBalancer* - 在当前云中创建一个外部负载均衡器(如果支持的话)，并为 Service 分配一个固定的外部IP。是 NodePort 的超集。
- *ExternalName* - 通过返回带有该名称的 CNAME 记录，使用任意名称(由 spec 中的`externalName`指定)公开 Service。不使用代理。这种类型需要`kube-dns`的v1.7或更高版本。

更多关于不同 Service 类型的信息可以在[使用源 IP](https://kubernetes.io/zh-cn/docs/tutorials/services/source-ip/) 教程。 也请参阅 [连接应用程序和 Service ](https://kubernetes.io/zh-cn/docs/concepts/services-networking/connect-applications-service)。

另外，需要注意的是有一些 Service 的用例没有在 spec 中定义`selector`。 一个没有`selector`创建的 Service 也不会创建相应的端点对象。这允许用户手动将服务映射到特定的端点。没有 selector 的另一种可能是你严格使用`type: ExternalName`来标记。

#### Service 和 Label

![img](images/module_04_services.svg)

Service 通过一组 Pod 路由通信。Service 是一种抽象，它允许 Pod 死亡并在 Kubernetes 中复制，而不会影响应用程序。在依赖的 Pod (如应用程序中的前端和后端组件)之间进行发现和路由是由Kubernetes Service 处理的。

Service 匹配一组 Pod 是使用 [标签(Label)和选择器(Selector)](https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/labels), 它们是允许对 Kubernetes 中的对象进行逻辑操作的一种分组原语。标签(Label)是附加在对象上的键/值对，可以以多种方式使用:

- 指定用于开发，测试和生产的对象
- 嵌入版本标签
- 使用 Label 将对象进行分类

![img](images/module_04_labels.svg)



标签(Label)可以在创建时或之后附加到对象上。他们可以随时被修改。现在使用 Service 发布我们的应用程序并添加一些 Label 。



总结：

- *Kubernetes 的 Service 是一个抽象层，它定义了一组 Pod 的逻辑集，并为这些 Pod 支持外部流量暴露、负载平衡和服务发现。*
- *你也可以在创建 Deployment 的同时用 `--expose`创建一个 Service 。*



### 扩缩应用程序

在之前的模块中，我们创建了一个 [Deployment](https://kubernetes.io/zh-cn/docs/concepts/workloads/controllers/deployment/)，然后通过 [Service](https://kubernetes.io/zh-cn/docs/concepts/services-networking/service/)让其可以开放访问。Deployment 仅为跑这个应用程序创建了一个 Pod。 当流量增加时，我们需要扩容应用程序满足用户需求。

**扩缩** 是通过改变 Deployment 中的副本数量来实现的。

#### 扩缩概述

![img](images/module_05_scaling1.svg)

扩展 Deployment 将创建新的 Pods，并将资源调度请求分配到有可用资源的节点上，收缩 会将 Pods 数量减少至所需的状态。Kubernetes 还支持 Pods 的[自动缩放](https://kubernetes.io/zh-cn/docs/tasks/run-application/horizontal-pod-autoscale/)，但这并不在本教程的讨论范围内。将 Pods 数量收缩到0也是可以的，但这会终止 Deployment 上所有已经部署的 Pods。

运行应用程序的多个实例需要在它们之间分配流量。服务 (Service)有一种负载均衡器类型，可以将网络流量均衡分配到外部可访问的 Pods 上。服务将会一直通过端点来监视 Pods 的运行，保证流量只分配到可用的 Pods 上。

*扩缩是通过改变 Deployment 中的副本数量来实现的。*



一旦有了多个应用实例，就可以没有宕机地滚动更新。我们将会在下面的模块中介绍这些。现在让我们使用在线终端来体验一下应用程序的扩缩过程。

总结：

- *在运行 kubectl run 命令时，你可以通过设置 --replicas 参数来设置 Deployment 的副本数。*
- *扩缩是通过改变 Deployment 中的副本数量来实现的。*

#### 执行滚动更新

#### 目标

- 使用 kubectl 执行滚动更新。

#### 更新应用程序

用户希望应用程序始终可用，而开发人员则需要每天多次部署它们的新版本。在 Kubernetes 中，这些是通过滚动更新（Rolling Updates）完成的。 **滚动更新** 允许通过使用新的实例逐步更新 Pod 实例，零停机进行 Deployment 更新。新的 Pod 将在具有可用资源的节点上进行调度。

在前面的模块中，我们将应用程序扩展为运行多个实例。这是在不影响应用程序可用性的情况下执行更新的要求。默认情况下，更新期间不可用的 pod 的最大值和可以创建的新 pod 数都是 1。这两个选项都可以配置为（pod）数字或百分比。 在 Kubernetes 中，更新是经过版本控制的，任何 Deployment 更新都可以恢复到以前的（稳定）版本。



#### 滚动更新概述

![img](images/module_06_rollingupdates1.svg)

与应用程序扩展类似，如果 Deployment 是公开的，服务将在更新期间仅对可用的 pod 进行负载均衡。可用 Pod 是应用程序用户可用的实例。

滚动更新允许以下操作：

- 将应用程序从一个环境提升到另一个环境（通过容器镜像更新）
- 回滚到以前的版本
- 持续集成和持续交付应用程序，无需停机

*如果 Deployment 是公开的，则服务将仅在更新期间对可用的 pod 进行负载均衡。*



在下面的交互式教程中，我们将应用程序更新为新版本，并执行回滚。

总结：

- *滚动更新允许通过使用新的实例逐步更新 Pod 实例从而实现 Deployments 更新，停机时间为零。*
- *如果 Deployment 是公开的，则服务将仅在更新期间对可用的 pod 进行负载均衡。*
