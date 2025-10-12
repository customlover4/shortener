package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)
var client = &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:     1000, // Увеличиваем в 10 раз
			MaxIdleConns:        500,  // Увеличиваем в 5 раз
			MaxIdleConnsPerHost: 100,  // КРИТИЧЕСКИ ВАЖНО!
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		Timeout: 5 * time.Second,
	}
func worker(wg *sync.WaitGroup) {
	defer wg.Done()

	req, err := http.NewRequest(http.MethodGet, "http://localhost/s/test2", nil)
	if err != nil {
		fmt.Println(err)
	}
	client.Do(req)
}

func main() {
	start := time.Now()
	wg := new(sync.WaitGroup)
	total := float64(0)
	for i := 0; i < 1; i++ {
		start := time.Now()
		for j := 0; j < 5; j++ {
			wg.Add(1)
			go worker(wg)
		}
		wg.Wait()
		end := time.Since(start)
		total += float64(end.Milliseconds()) / 100
	}
	diff := time.Since(start).Milliseconds()
	fmt.Println("total time: ", diff, "ms")
	fmt.Println("total radio: ", total/100, "ms")
}

// 10 workers batchsize = 500 maxconn = 100 idleconn = 20
// total radio:  0.8720900000000033 ms
// total radio:  1.209 ms

// 10 workers batchsize = 10000 maxconn = 100 idleconn = 20
// total radio:  0.07530000000000006 ms
// total radio:  0.4281000000000001 ms
// total radio:  1.1844 ms
