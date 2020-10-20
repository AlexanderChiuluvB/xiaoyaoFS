BINARY = xiaoyao

SOURCE_DIR = /

all: debug_mount

debug_mount:
	go build -gcflags="all=-N -l"
	dlv --listen=:2345 --accept-multiclient exec xiaoyao -- mount -dir=~/tmp/mm -filer.path=/buckets