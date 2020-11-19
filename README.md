# xiaoyaoFS
逍遥FS 本科毕业设计

## 开发记录

### TODO 2020/11/15

1.storage server多副本一致性问题


2.storage server/master server的缓存机制


3.FUSE的目录实现，缓存机制


### 下一步计划

1. docker 部署Hbase2.0的尝试，主要是为了用Accordion特性, 之前测试都是用1.26

2. go的hbase 客户端写的真烂，没有批量读取功能，需要自己开gorouine实现

3. 在master层访问hbase获得metadata的时候，加一层布隆过滤器可以用blocked bloom filter

4. 在master尝试不同的数据库来存meta,可以试试redis,badgerDB,tikv 

