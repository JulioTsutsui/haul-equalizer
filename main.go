package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	backendServers = []string{
		"http://localhost:9001",
		"http://localhost:9002",
		"http://localhost:9003",
		"http://localhost:9004",
	}
	backendHealth = make(map[string]bool)
	currentIndex  = 0
	mu            sync.Mutex
	maxRetries    = 3
)

func healthCheck() {
	for {
		for _, server := range backendServers {
			resp, err := http.Get(server)
			if err != nil || resp.StatusCode >= 500 {
				backendHealth[server] = false
				log.Printf("Server Down: %s", server)
			} else {
				backendHealth[server] = true
			}

			if resp != nil {
				resp.Body.Close()
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func seedBackendServers(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received at server on port %s", port)
		fmt.Fprintf(w, "Response from server on port: %s\n", port)
	})

	go func() {
		log.Printf("Starting server on port %s", port)
		err := http.ListenAndServe(":"+port, mux)
		if err != nil {
			log.Fatalf("Error starting server on port: %s:%v", port, err)
		}
	}()
}

func HaulEqualizer(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < maxRetries; i++ {
		mu.Lock()
		startIndex := currentIndex

		for {
			targetUrl := backendServers[currentIndex]
			currentIndex = (currentIndex + 1) % len(backendServers)

			if backendHealth[targetUrl] {
				mu.Unlock()
				proxyReq, err := http.NewRequest(r.Method, targetUrl, r.Body)
				if err != nil {
					log.Printf("Failed to create a request to %s: %v", targetUrl, err)
					continue
				}

				proxyReq.Header = r.Header.Clone()

				client := http.Client{
					Transport: &http.Transport{
						MaxIdleConnsPerHost: 10,
					},
				}

				proxyRes, err := client.Do(proxyReq)
				if err != nil {
					log.Printf("Failed to proxy request to backend %s: %v", targetUrl, err)
					continue
				}

				defer proxyRes.Body.Close()

				for key, values := range proxyRes.Header {
					for _, value := range values {
						w.Header().Set(key, value)
					}
				}

				w.WriteHeader(proxyRes.StatusCode)

				// Stream the body back to client
				_, err = io.Copy(w, proxyRes.Body)
				if err != nil {
					log.Printf("Error copying response body: %v", err)
					continue
				}

				return
			}

			if currentIndex == startIndex {
				break
			}

		}

		mu.Unlock()
	}

	w.WriteHeader(http.StatusServiceUnavailable)
	fmt.Fprintf(w, "No healthy servers available, try again later \n")
}

func main() {
	lb_port := "9000"

	backend_ports := []string{
		"9001",
		"9002",
		"9003",
	}

	for _, port := range backend_ports {
		backendHealth["http://localhost:"+port] = true
		seedBackendServers(port)
	}

	// wait for servers to setup properly before check the health
	time.Sleep(5 * time.Second)

	go healthCheck()

	mux := http.NewServeMux()
	mux.HandleFunc("/", HaulEqualizer)

	log.Println("Load Balancer running on port", lb_port)

	err := http.ListenAndServe(":"+lb_port, mux)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
