docker exec -it xxx 
cqlsh 进入命令行

 


1. create a keyspace

CREATE KEYSPACE xiaoyaofs WITH replication = {'class':'SimpleStrategy', 'replication_factor' : 1};

2. create filemeta table

 USE xiaoyaofs;

 CREATE TABLE metadata (
    filePath varchar,
    vid bigint,
    nid bigint,
    PRIMARY KEY (vid, nid)
 ) WITH CLUSTERING ORDER BY (nid ASC);


 
