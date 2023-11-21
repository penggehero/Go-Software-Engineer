# 服务治理之负载均衡



## 客户端负载均衡：

### ConsistentHash

下面摘抄一下kitex 官网的说明

一致性哈希主要适用于对上下文（如实例本地缓存）依赖程度高的场景，如希望同一个类型的请求打到同一台机器，则可使用该负载均衡方法。

**如果你不了解什么是一致性哈希，或者不知道带来的副作用，请勿使用一致性哈希。**



### WeightedRandom

基于权重的随机策略。

 Kitex 的默认策略，会依据实例的权重进行加权随机，并保证每个实例分配到的负载和自己的权重成比例。



### Weighted round robin





### Power of two choices



### Random



### Round-robin



### Least connection