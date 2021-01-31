package main

import (
	"fmt"
	"github.com/AlexanderChiuluvB/xiaoyaoFS/master/api"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type result struct {
	concurrent  int
	num         int
	startTime   time.Time
	endTime     time.Time
	completed   int32
	failed      int32
	transferred uint64
}

func main() {
	const CONCURRENCY = 16 //allowed
	var totalSize int64
	var uploadedSize int64
	var readSize int64
	var jpgList []string
	sizeMap := make(map[string]int64)
	//Open Directory, get all pic list
	err := filepath.Walk("/home/hadoop/Downloads/Panorama", func (path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "jpg") {
			jpgList = append(jpgList, path)
			totalSize += info.Size()
			sizeMap[path] = info.Size()
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}

	uploadResult := &result{
		concurrent: CONCURRENCY,
		startTime: time.Now(),
	}

	loop := make(chan string)
	wg := sync.WaitGroup{}

	for i:=0; i < uploadResult.concurrent; i ++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range loop {
				err := api.Upload("localhost", 8888, path, path)
				if err == nil {
					atomic.AddInt32(&uploadResult.completed, 1)
					uploadedSize += sizeMap[path]
				} else {
					atomic.AddInt32(&uploadResult.failed, 1)
					fmt.Println("upload failed ", err.Error())
				}
			}
		}()
	}

	for _, path := range jpgList {
		loop <- path
	}

	close(loop)
	wg.Wait()
	uploadResult.endTime = time.Now()
	timeTaken := float64(uploadResult.endTime.UnixNano() - uploadResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("upload %d %dbyte file:\n\n", len(jpgList), totalSize)
	fmt.Printf("concurrent:             %d\n", uploadResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", uploadResult.completed)
	fmt.Printf("failed:                 %d\n", uploadResult.failed)
	fmt.Printf("transferred:            %d byte\n", uploadedSize)
	fmt.Printf("request per second:     %.2f\n", float64(uploadResult.completed) / timeTaken)
	fmt.Printf("transferred per second: %.2f MB/s\n", float64(uploadedSize)/timeTaken/(1024*1024))


	readResult := &result{
		concurrent: CONCURRENCY,
		startTime: time.Now(),
	}
	loop = make(chan string)

	for i:=0; i < readResult.concurrent; i ++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range loop {
				file, _ := os.Stat(path)
				data, err := api.Get("localhost", 8888, path)
				if err == nil && int64(len(data)) == file.Size() {
					atomic.AddInt32(&readResult.completed, 1)
					readSize += sizeMap[path]
				} else {
					atomic.AddInt32(&readResult.failed, 1)
					fmt.Println("Read failed ", err.Error())
				}
			}
		}()
	}
	for _, path := range jpgList {
		loop <- path
	}

	close(loop)
	wg.Wait()

	readResult.endTime = time.Now()
	timeTaken = float64(readResult.endTime.UnixNano() - readResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("read %d file, total %dbyte file:\n\n", len(jpgList), totalSize)
	fmt.Printf("concurrent:             %d\n", readResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", readResult.completed)
	fmt.Printf("failed:                 %d\n", readResult.failed)
	fmt.Printf("transferred:            %d byte\n", readSize)
	fmt.Printf("request per second:     %.2f\n", float64(readResult.num) / timeTaken)
	fmt.Printf("transferred per second: %.2f MB/s \n", float64(readSize)/timeTaken/(1024*1024))


	deleteResult := &result{
		concurrent: CONCURRENCY,
		startTime: time.Now(),
	}
	loop = make(chan string)

	for i:=0; i < deleteResult.concurrent; i ++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range loop {
				err := api.Delete("localhost", 8888, path)
				if err == nil {
					atomic.AddInt32(&deleteResult.completed, 1)
					readSize += sizeMap[path]
				} else {
					atomic.AddInt32(&deleteResult.failed, 1)
					fmt.Println("upload failed ", err.Error())
				}
			}
		}()
	}

	for _, path := range jpgList {
		loop <- path
	}
	close(loop)
	wg.Wait()

	deleteResult.endTime = time.Now()
	timeTaken = float64(deleteResult.endTime.UnixNano() - deleteResult.startTime.UnixNano()) / float64(time.Second)

	fmt.Printf("delete%d file:\n\n", len(jpgList))
	fmt.Printf("concurrent:             %d\n", deleteResult.concurrent)
	fmt.Printf("time taken:             %.2f seconds\n", timeTaken)
	fmt.Printf("completed:              %d\n", deleteResult.completed)
	fmt.Printf("failed:                 %d\n", deleteResult.failed)
	//fmt.Printf("transferred:            %d byte\n", readSize)
	fmt.Printf("request per second:     %.2f\n", float64(deleteResult.completed) / timeTaken)
	//fmt.Printf("transferred per second: %.2f byte\n", float64(readSize)/timeTaken)


}



