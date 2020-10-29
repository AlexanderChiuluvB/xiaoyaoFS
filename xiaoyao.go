package main

import (
	"github.com/AlexanderChiuluvB/xiaoyaoFS/storage"
)

func main() {
	var (
		config *storage.Config
		store *storage.Store
		err error
	)
	if config, err = storage.NewConfig("./storage/store.toml"); err != nil {
		panic(err)
	}

	if store, err = storage.NewStore(config); err != nil {
		panic(err)
	}

	store.Start()
}


