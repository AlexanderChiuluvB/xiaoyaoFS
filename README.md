# xiaoyaoFS
逍遥FS 本科毕业设计

## 开发记录

### TODO 2020/11/15

1.storage server多副本一致性问题


2.storage server/master server的缓存机制


3.FUSE的目录实现，缓存机制

### TODO 2020/12/1




### API 使用


上传 
```
curl -F file=@localFilePath  'http://localhost:8888/uploadFile?filepath=/example.png'
```

获取
```
wget 'http://localhost:8888/getFile?filepath=/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/test/nut.png'
```

删除metadata(不会删除Volume中的真实数据)
```
curl -X DELETE 'http://localhost:8888/deleteFile?filepath=/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/test/nut.png'
```


### 分支说明

#### main 分支

每个Volume文件都由一个嵌入式的leveldb维护KEY-VALUE映射关系

Master Server 用Redis 存储 <FileName, <Vid,Nid>> 的映射关系

Storage Server 用LevelDB 存储<<Vid,Nid>, Needle> 的映射关系

master.toml 

```
StoreDir = "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/masterDir"

MasterHost = "localhost"
MasterPort = 8888

MetaType = "Redis"
# redis
RedisHost = "localhost"
RedisPort = 6379
#Password = "110120"
Database = 0

MaxVolumeNum = 5
```
storage.toml

```
StoreDir = "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/storeDir1"
MountDir = "/Users/alex/mountTest1"
MasterHost = "localhost"
MasterPort = 8888
StoreApiHost = "localhost"
StoreApiPort = 7900

```

docker 启动 redis

```
docker run -p 6379:6379 -d redis:latest redis-server
```

单Master单Storage启动方法


```
./xiaoyaoFS --config=master.toml master
./xiaoyaoFS --config=store1.toml storage
```
### 挂载

切换到FUSE 分支。FUSE 分支支持在Master Server使用Redis/Hbase/Cassandra存储metadata。默认使用redis

挂载到某文件夹
具体文件夹路径在配置文件中设置MountDir = "/Users/alex/mountTest1"

```
./xiaoyaoFS --config=store1.toml mount
```

以下分支都不支持挂载

#### leveldb_one_storage_no_entry

和main分支区别为单个LevelDB维护整个Storage Server所有Volume对应的Key value关系

Master Server 用Redis/LevelDB 存储 <FileName, <Vid,Nid>> 的映射关系

Storage Server 用LevelDB 存储<<Vid,Nid>, Needle> 的映射关系

#### clickhouse_no_entry

在db/clickhouse 中,`docker compose up -d`启动一个单节点ClickHouse做测试

storage Server使用ClickHouse维护<<Vid,Nid>, Needle> 的映射关系

#### cassandra_no_entry
在db/cassandra中，用,`docker compose up -d`启动一个双节点集群，并且参照README.md创建表

storage Server使用Cassandra维护<<Vid,Nid>, Needle> 的映射关系

#### badger_one_storage_no_entry

master Server/storage Server使用Badger维护<<Vid,Nid>, Needle> 的映射关系


