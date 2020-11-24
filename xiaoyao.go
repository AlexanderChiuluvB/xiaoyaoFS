package main

import (
	"fmt"
	fuse2 "github.com/AlexanderChiuluvB/xiaoyaoFS/mount"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/config"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/utils/parser"
	b "github.com/AlexanderChiuluvB/xiaoyaoFS/utils/benchmark"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/seaweedfs/fuse"
	"github.com/seaweedfs/fuse/fs"
	"os"
	"os/signal"
	"syscall"
)

var (
	app = kingpin.New("whalefs", "A simple filesystem for small file.")
	configFile = app.Flag("config", "config file(toml)").Required().String()
	masterServer = app.Command("master", "master server")
	storageServer = app.Command("storage", "storage server")
	mount = app.Command("mount", "mount the xiaoyaoFs to a directory")
	benchmark = app.Command("benchmark", "benchmark")
	bmMasterHost = benchmark.Flag("masterHost", "host of master server").Default("localhost").String()
	bmMasterPort = benchmark.Flag("masterPort", "post of master server").Default("8888").Int()
	bmConcurrent = benchmark.Flag("concurrent", "concurrent").Default("16").Int()
	bmNum = benchmark.Flag("num", "number of file write/read").Default("1").Int()
	bmSize = parser.Size(benchmark.Flag("size", "size of file write/read").Default("1024B"))
	)

func main() {
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch command {
	case masterServer.FullCommand():
		startMaster(*configFile)
	case storageServer.FullCommand():
		startStorageServer(*configFile)
	case mount.FullCommand():
		startMount(*configFile)
	case benchmark.FullCommand():
		b.Benchmark(*bmMasterHost, *bmMasterPort, *bmConcurrent, *bmNum, int(*bmSize))
	}

}

func startMount(configFile string) {
	c, err := config.NewConfig(configFile)
	if err != nil {
		panic(fmt.Errorf("NewConfig(\"%s\") error(%v)", configFile, err))
	}

	options := []fuse.MountOption {
		fuse.VolumeName("xiaoyaoFS"),
		fuse.LocalVolume(),
	}

	conn, err := fuse.Mount(c.MountDir, options...)
	if err != nil {
		panic(fmt.Errorf("mount %s error %v", c.MountDir, err))
	}
	defer fuse.Unmount(c.MountDir)

	xiaoyaoFileSystem := fuse2.NewXiaoyaoFs(c)

	err = fs.Serve(conn, xiaoyaoFileSystem)

	<-conn.Ready
	if err := conn.MountError; err != nil {
		panic(fmt.Errorf("mount process: %v", err))
	}
}

func startStorageServer(configFile string) {
	c, err := config.NewConfig(configFile)
	if err != nil {
		panic(fmt.Errorf("NewConfig(\"%s\") error(%v)", configFile, err))
	}
	ss, err := storage.NewStore(c)
	if err != nil {
		panic(fmt.Errorf("NewStore(\"%s\") error(%v)", configFile, err))
	}
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		ss.Close()
		switch sig {
		case syscall.SIGINT:
			os.Exit(int(syscall.SIGINT))
		case syscall.SIGTERM:
			os.Exit(int(syscall.SIGTERM))
		}
	}()
	ss.Start()
}

func startMaster(configFile string) {
	c, err := config.NewConfig(configFile)
	if err != nil {
		panic(fmt.Errorf("NewConfig(\"%s\") error(%v)", configFile, err))
	}
	m, err := master.NewMaster(c)
	if err != nil {
		panic(fmt.Errorf("NewMaster(\"%s\") error(%v)", configFile, err))
	}


	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		m.Close()
		switch sig {
		case syscall.SIGINT:
			os.Exit(int(syscall.SIGINT))
		case syscall.SIGTERM:
			os.Exit(int(syscall.SIGTERM))
		}
	}()

	m.Start()
}





