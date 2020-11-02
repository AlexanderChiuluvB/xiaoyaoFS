package main

import (
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/signal"
	"syscall"
)

var (
	app = kingpin.New("whalefs", "A simple filesystem for small file.")
	configFile = app.Flag("config", "config file(toml)").Required().String()
	masterServer = app.Command("master", "master server")
	storageServer = app.Command("storage", "storage server")
)

func main() {
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	switch command {
	case masterServer.FullCommand():
		startMaster(*configFile)
	case storageServer.FullCommand():
		startStorageServer(*configFile)
	}

}

func startStorageServer(configFile string) {
	c, err := storage.NewConfig(configFile)
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
	c, err := storage.NewConfig(configFile)
	if err != nil {
		panic(fmt.Errorf("NewConfig(\"%s\") error(%v)", configFile, err))
	}
	m, err := master.NewMaster(c)
	if err != nil {
		panic(fmt.Errorf("NewMaster(\"%s\") error(%v)", configFile, err))
	}
	m.Metadata, err = master.NewHbaseStore(c)
	if err != nil {
		panic(fmt.Errorf("NewHbaseStore error %v", err))
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





