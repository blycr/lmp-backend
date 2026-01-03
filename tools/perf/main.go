package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	target := flag.String("target", "http://localhost:8080", "base url")
	duration := flag.Duration("duration", 10*time.Second, "test duration")
	concurrency := flag.Int("concurrency", 20, "number of workers")
	flag.Parse()

	client := &http.Client{Timeout: 5 * time.Second}
	var total, ok uint64
	var min, max int64
	var sum int64
	min = int64(^uint64(0) >> 1)

	stop := time.Now().Add(*duration)
	var wg sync.WaitGroup
	wg.Add(*concurrency)
	for i := 0; i < *concurrency; i++ {
		go func() {
			defer wg.Done()
			for time.Now().Before(stop) {
				start := time.Now()
				resp, err := client.Get(*target + "/health")
				lat := time.Since(start).Milliseconds()
				atomic.AddUint64(&total, 1)
				if err == nil && resp.StatusCode == 200 {
					atomic.AddUint64(&ok, 1)
					atomic.AddInt64(&sum, lat)
					for {
						old := atomic.LoadInt64(&min)
						if lat < old {
							if atomic.CompareAndSwapInt64(&min, old, lat) {
								break
							}
						} else {
							break
						}
					}
					for {
						old := atomic.LoadInt64(&max)
						if lat > old {
							if atomic.CompareAndSwapInt64(&max, old, lat) {
								break
							}
						} else {
							break
						}
					}
				}
				if resp != nil && resp.Body != nil {
					resp.Body.Close()
				}
			}
		}()
	}
	wg.Wait()
	avg := float64(0)
	if ok > 0 {
		avg = float64(sum) / float64(ok)
	}
	rps := float64(ok) / duration.Seconds()
	fmt.Printf("total=%d ok=%d rps=%.2f min_ms=%d avg_ms=%.2f max_ms=%d\n", total, ok, rps, min, avg, max)
}

