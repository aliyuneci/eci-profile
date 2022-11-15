## Introduction
我们基于 Kubernetes 的 Mutating Webhook 机制实现了 ECI-Profile，它将为那些调度到虚拟节点上的 Pod 自动追加 Annotations、Labels、Tolerations 等，简化了线下 IDC 集群无缝对接阿里云 [ECI](https://www.aliyun.com/product/eci) 的流程。

## Feature
在混合云场景下，用户的集群中包括了普通节点和虚拟节点，假如我们想把特定的 Pod 调度到虚拟节点上，需要对 Pod 的 Spec 清单进行修改，无法做到零侵入。另一点就是如果想将大量 Pod 调度到虚拟节点上，需要对每一个 Pod 的 Spec 清单进行修改，这是不现实的。而 ECI-Profile 解决了这个问题，您可以声明需要匹配的 Pod Labels，对于 Labels 能够匹配上的 Pod，根据预设的调度策略，将被自动追加已配置的 Tolerations，协调 Kube-Scheduler 的调度。甚至还可以声明需要匹配的 Namespace Labels，对于 Labels 能够匹配上的 Namespace，在该 Namespace 下创建的 Pod 都将被影响。预设的调度策略包括：公平策略，追加虚拟节点的 Tolerations；仅调度普通节点策略，不追加任何 Tolerations;普通节点优先策略，当普通节点调度失败时才追加虚拟节点的 Tolerations;仅调度到虚拟节点策略，追加虚拟节点的 Tolerations 和 NodeSelector。除此之外，调度到虚拟节点的 Pod 是具有一些阿里云相关功能特性的，如指定 ECS 实例规格，启用镜像缓存，设置 NTP 服务等，一般需要在 Pod 中追加 Annotations 或者 Labels 来实现这些功能特性的开启或关闭。ECI-Profile 可实现自动追加 Annotations 和 Labels 的功能，和之前的调度一样，您可以声明需要匹配的 Pod Labels，以及需要追加的 Annotations 和 Labels，对于 Labels 能够匹配上的 Pod，将被自动追加配置的 Annotations 和 Labels。同样的，您也可以声明需要匹配的 Namespace Labels，对于 Labels 能够匹配上的 Namespace，在该 Namespace 下创建的 Pod 都将被影响。

## Build
> docker build -t xxx/eci-profile:latest
> docker push xxx/eci-profile:latest

## Deploy
在 k8s 集群中部署 ECI-Profile
> kubectl apply -f deploy.yaml

## Example
ECI-Profile 可以通过 Pod/Namespace 的 Labels 筛选符合条件的 Pod，完成以下功能：

#### 注入 Annotations/Labels
为调度到虚拟节点上的 Pod 绑定阿里云 EIP。关于 ECI Pod Annotations 的更多信息，请参考[链接](https://help.aliyun.com/document_detail/144561.html)。
```yaml
apiVersion: eci.aliyun.com/v1beta1
kind: Selector
metadata:
  name: test-fair
spec:
  objectLabels:
    matchLabels:
      app: nginx
  effect:
    annotations:
      k8s.aliyun.com/eci-with-eip: "true"
  policy:
    fair: {}
  priority: 3 # priority 表示优先级，当集群中存在多个 Selector 时，优先级最高的 Selector 将会被应用。
```
优先删除虚拟节点上的 Pod，需要 Kubernetes 版本为 1.22 及以上。更多介绍请参考[链接](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/#pod-deletion-cost)
```yaml
apiVersion: eci.aliyun.com/v1beta1
kind: Selector
metadata:
  name: test-fair
spec:
  objectLabels:
    matchLabels:
      app: nginx
  effect:
    annotations:
      controller.kubernetes.io/pod-deletion-cost: -1000
  policy:
    fair: {}
  priority: 3 # priority 表示优先级，当集群中存在多个 Selector 时，优先级最高的 Selector 将会被应用。
```

#### 执行调度策略
公平调度（fair），为选中的 Pod 增加虚拟节点容忍，由 Kube-Scheduler 决定调度。
```yaml
apiVersion: eci.aliyun.com/v1beta1
kind: Selector
metadata:
  name: test-fair
spec:
  objectLabels:
    matchLabels:
      app: nginx
  effect:
    annotations:
      foo: boo
    labels:
      foo: boo 
  policy:
    fair: {}
  priority: 3 # priority 表示优先级，当集群中存在多个 Selector 时，优先级最高的 Selector 将会被应用。
```
标准节点优先（normalNodePrefer）：标准节点资源不足时允许调度到虚拟节点。
```yaml
apiVersion: eci.aliyun.com/v1beta1
kind: Selector
metadata:
  name: test-normal-node-prefer
spec:
  objectLabels:
    matchLabels:
      app: nginx
  effect:
    annotations:
      foo: boo
    labels:
      foo: boo
  policy:
    normalNodePrefer: {}
  # priority: 3 # priority 表示优先级，当集群中存在多个 Selector 时，优先级最高的 Selector 将会被应用。
```
仅调度到虚拟节点（virtualNodeOnly）：为选中的 Pod 增加虚拟节点容忍及虚拟节点的 NodeSelector，Pod 只会调度到虚拟节点。
```yaml
apiVersion: eci.aliyun.com/v1beta1
kind: Selector
metadata:
  name: test-virtual-node-only
spec:
  objectLabels:
    matchLabels:
      app: nginx
  effect:
    annotations:
      foo: boo
    labels:
      foo: boo
  policy:
    virtualNodeOnly: {}
  # priority: 2 # priority 表示优先级，当集群中存在多个 Selector 时，优先级最高的 Selector 将会被应用。
```