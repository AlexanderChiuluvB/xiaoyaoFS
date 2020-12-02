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

建议生产上使用该分支，并且Master 默认使用 ClickHouse 存储映射关系

单个LevelDB维护整个Storage Server所有Volume对应的Key value关系

Master Server 用Redis/LevelDB 存储 <FileName, <Vid,Nid>> 的映射关系

Storage Server 用LevelDB 存储<<Vid,Nid>, Needle> 的映射关系


master.toml 

```
StoreDir = "/Users/alex/go/src/github.com/AlexanderChiuluvB/xiaoyaoFS/masterDir"

MasterHost = "localhost"
MasterPort = 8888

MetaType = "ClickHouse"

# 缓存超时时间 5min
ExpireTime = "5m"
# 清除超时数据的周期 10min
PurgeTime = "10m"

# ClickHouse
ClickHouseHost = "tcp://127.0.0.1:9000?debug=true"

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

StoreApiHost = "localhost"
StoreApiPort = 7900
MasterHost = "localhost"
MasterPort = 8888

MountDir = "/Users/alex/mountTest1"
```

docker 启动 单节点ClickHouse
```
cd db/clickhouse
docker-compose up -d
```

如果要用Redis,把master.toml的MeteType改为Redis,并且调整相关参数RedisHost, RedisPort
```
docker run -p 6379:6379 -d redis:latest redis-server
```

单Master单Storage启动方法
```
./xiaoyaoFS --config=master.toml master
./xiaoyaoFS --config=store1.toml storage
```

单Master多Storage
```
./xiaoyaoFS --config=master.toml master
./xiaoyaoFS --config=store1.toml storage
./xiaoyaoFS --config=store2.toml storage
```
注意store1.toml和store2.toml的storeDir挂载路径和StoreApiHost,StoreApiPort都应该做相应调整。

### 挂载

切换到FUSE 分支。FUSE 分支支持在Master Server使用Redis/Hbase/Cassandra/ClickHouse存储metadata。默认使用redis
但是Redis重启后所有数据将丢失，不能持久化。所以开发计划是默认用clickhouse

挂载到某文件夹
具体文件夹路径在配置文件中设置MountDir = "/Users/alex/mountTest1"

启动了master和storage server后:
```
./xiaoyaoFS --config=store1.toml mount
```

然后上传文件的时候，必须要保证filepath参数和挂载路径相同

curl -F file=@localFilePath  'http://localhost:8888/uploadFile?filepath=/Users/alex/mountTest1/example.png'

如果上传的是目录结构，那么需要手动在挂载的文件夹创建相应的目录，如

curl -F file=@localFilePath  'http://localhost:8888/uploadFile?filepath=/Users/alex/mountTest1/testdir/example.png'

那么首先要手动在mountTest1文件夹创建testdir文件夹

### 其他开发分支

#### clickhouse_no_entry

在db/clickhouse 中,`docker compose up -d`启动一个单节点ClickHouse做测试

storage Server使用ClickHouse维护<<Vid,Nid>, Needle> 的映射关系

#### cassandra_no_entry
在db/cassandra中，用,`docker compose up -d`启动一个双节点集群，并且参照README.md创建表

storage Server使用Cassandra维护<<Vid,Nid>, Needle> 的映射关系

#### badger_one_storage_no_entry

master Server/storage Server使用Badger维护<<Vid,Nid>, Needle> 的映射关系

