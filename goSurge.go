package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

type RequestResult struct {
	ResponseCode int
	RequestTime  time.Duration
}

type Statistics struct {
	TotalRequests     int
	FailedRequests    int
	RequestsPerSecond float64

	TotalRequestTime struct {
		Min  time.Duration
		Max  time.Duration
		Mean time.Duration
	}

	TimeToFirstByte struct {
		Min  time.Duration
		Max  time.Duration
		Mean time.Duration
	}

	TimeToLastByte struct {
		Min  time.Duration
		Max  time.Duration
		Mean time.Duration
	}
}

///declare a wait group globally

var wg sync.WaitGroup

//function to calculate statistics

func calculateStatistics(results chan RequestResult) Statistics {
	var (
		totalRequests    int
		failedRequests   int
		totalRequestTime time.Duration
		// totalFirstByte   time.Duration
		// totalLastByte    time.Duration
	)

	minRequestTime := time.Duration(int64(^uint64(0) >> 1)) // Max int64 value for initialization
	maxRequestTime := time.Duration(0)

	minFirstByte := time.Duration(int64(^uint64(0) >> 1))
	maxFirstByte := time.Duration(0)

	minLastByte := time.Duration(int64(^uint64(0) >> 1))
	maxLastByte := time.Duration(0)

	for result := range results {
		totalRequests++

		if result.ResponseCode < 200 && result.ResponseCode > 299 {
			failedRequests++
		}

		// Update total request time
		totalRequestTime += result.RequestTime

		// Update time to first byte
		if result.RequestTime < minFirstByte {
			minFirstByte = result.RequestTime
		}
		if result.RequestTime > maxFirstByte {
			maxFirstByte = result.RequestTime
		}
		if result.RequestTime < minRequestTime {
			minRequestTime = result.RequestTime
		}
		if result.RequestTime > maxRequestTime {
			maxRequestTime = result.RequestTime
		}
		// Simulate processing the request and obtaining the last byte time
		lastByteTime := result.RequestTime / 2

		// Update time to last byte
		if lastByteTime < minLastByte {
			minLastByte = lastByteTime
		}
		if lastByteTime > maxLastByte {
			maxLastByte = lastByteTime
		}
	}

	// Calculate mean values
	meanRequestTime := totalRequestTime / time.Duration(totalRequests)
	meanFirstByte := (minFirstByte + maxFirstByte) / 2
	meanLastByte := (minLastByte + maxLastByte) / 2

	return Statistics{
		TotalRequests:     totalRequests,
		FailedRequests:    failedRequests,
		RequestsPerSecond: float64(totalRequests) / 10, // Adjust 10 to the actual concurrency value
		TotalRequestTime:  struct{ Min, Max, Mean time.Duration }{minRequestTime, maxRequestTime, meanRequestTime},
		TimeToFirstByte:   struct{ Min, Max, Mean time.Duration }{minFirstByte, maxFirstByte, meanFirstByte},
		TimeToLastByte:    struct{ Min, Max, Mean time.Duration }{minLastByte, maxLastByte, meanLastByte},
	}
}

// function to make request for http
func makeRequest(u string, method string, contentType *string, content string, responseBody bool, headers bool, n int) int {

	jsonBody := []byte(content)
	bodyReader := bytes.NewReader(jsonBody)
	// fmt.Println(json.Valid(jsonBody))
	// fmt.Println(bodyReader)
	// if err != nil {
	// 	panic(err)
	// }
	req, err := http.NewRequest(method, u, bodyReader)
	req.Header.Set("Content-Type", *contentType)
	// fmt.Println(req, err)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}
	// fmt.Printf("client: status code: %d\n", res.StatusCode)
	if responseBody {
		if n < 2 {
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Printf("\nclient: could not read response body: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf("\nclient: response body: %s\n", resBody)
		} else {
			color.Red("\nthe number of requests should be less than 1 for getting response body")
			// panic(err)
			os.Exit(1)
		}
	}
	defer res.Body.Close()

	return res.StatusCode

}

func maketlsRequest(url string, tlsConfig *tls.Config) {
	// Create a custom HTTP client with the specified TLS configuration
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Make an HTTP GET request
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Process the response as needed
	fmt.Printf("Response code: %d\n", resp.StatusCode)
	// You can gather more stats here
}

func tlsVersionToString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}

func verifyTls(tcpAddr *net.TCPAddr, u *url.URL) {

	tlsConfigurations := []*tls.Config{
		&tls.Config{MinVersion: tls.VersionTLS10, MaxVersion: tls.VersionTLS10},
		&tls.Config{MinVersion: tls.VersionTLS11, MaxVersion: tls.VersionTLS11},
		&tls.Config{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12},
		&tls.Config{MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13},
	}

	for _, tlsConfig := range tlsConfigurations {
		fmt.Printf("Testing SSL/TLS Version: %s\n", tlsVersionToString(tlsConfig.MinVersion))

		start := time.Now()
		maketlsRequest(u.String(), tlsConfig)
		fmt.Printf("Request took: %v\n", time.Since(start))

		fmt.Println()
	}

}

func main() {

	myFigure := figure.NewColorFigure("GoSurge", "big", "green", true)
	boldGreen := color.New(color.FgGreen, color.Bold)
	// boldGreen.Println("This is bold and green")
	// boldText := color.New(color.Bold).SprintFunc()
	boldGreen.Println(myFigure)
	boldGreen.Println("  A simple HTTP Benchmarking and LoadTesting tool written in Go")

	method := flag.String("m", "GET", "`method`")
	urlString := flag.String("url", "", "url example : 'http://localhost:8080 or https://websitename.com'")
	contentType := flag.String("con-type", "text/plain", "Content type")
	help := flag.Bool(("help"), false, "help")
	content := flag.String("content", "", "Content for POST or PUT or GET or DELETE")
	responseBody := flag.Bool("res-body", false, "Response body")
	reqNum := flag.Int("n", 1, "Number of requests")
	conqreq := flag.Int("c", 1, "Number of concurrent requests")
	verTls := flag.Bool("tls", false, "Verify TLS configuration for the specified URL")
	headers := flag.Bool("H", false, "Headers for making request --type:boolean --default:false")
	filePath := flag.String("f", "", "file path containing urls as text file")

	flag.Parse()
	// fmt.Printf("this %s\n", method)
	if len(os.Args) < 1 {
		fmt.Println("Usage: gosurge <URL>")
		flag.PrintDefaults()
		// os.Exit(1)
	}
	// fmt.Println(*method, *urlString)

	if *help {
		fmt.Println("Usage: gosurge  <URL>")
		flag.PrintDefaults()
	}
	u, err := url.Parse(*urlString)
	if err != nil {
		panic(err)
	}
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "80" // Default port for HTTP
	}
	// fmt.Println(host)
	serverAddr := fmt.Sprintf("%s:%s", host, port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	// typ := reflect.TypeOf(u)
	// fmt.Printf("Type of %v:", tcpAddr)
	if err != nil {
		panic(err)
	}

	requestResults := make(chan RequestResult, *reqNum)

	conqreqChan := make(chan struct{}, *conqreq)

	done := make(chan struct{})

	go func() {
		displayLoadingAnimation(3 * time.Second) // Display for 3 seconds
		close(done)                              // Signal completion
	}()

	if *filePath != "" {
		urls, err := consuneUrlsfromFiles(*filePath)
		conqreqChan2 := make(chan struct{}, len(urls)**reqNum)
		fmt.Println(len(urls))
		if err != nil {
			panic(err)
		}
		for _, url := range urls {
			url := url
			// fmt.Println(i) // Create a new variable inside the loop and assign the value of url to it
			for i := 0; i < *reqNum; i++ {
				wg.Add(1)
				conqreqChan2 <- struct{}{}

				go func(u string) {
					defer wg.Done()
					start := time.Now()
					ResponseCode := makeRequest(u, *method, contentType, *content, *responseBody, *headers, *reqNum)
					// fmt.Println("Request took: ", ResponseCode)
					requestResults <- RequestResult{ResponseCode: ResponseCode, RequestTime: time.Since(start)}
					<-conqreqChan2
				}(url)
			}

		}
		go func() {
			wg.Wait()
			close(conqreqChan2)
		}()
	} else {

		if *urlString == "" {
			fmt.Println("Error: Server address is required.")
			flag.PrintDefaults()
			os.Exit(1)
		}

		for i := 0; i < *reqNum; i++ {
			wg.Add(1)
			conqreqChan <- struct{}{}

			go func() {
				defer wg.Done()
				start := time.Now()
				ResponseCode := makeRequest(*urlString, *method, contentType, *content, *responseBody, *headers, *reqNum)
				requestResults <- RequestResult{ResponseCode: ResponseCode, RequestTime: time.Since(start)}
				<-conqreqChan
			}()

		}
		go func() {
			wg.Wait()
			close(conqreqChan)
		}()
	}

	go func() {
		wg.Wait()
		close(requestResults)
		// close(done)
	}()
	stats := calculateStatistics(requestResults)
	displayResults(stats)

	if *verTls {
		fmt.Println("\n\n\rVerifying TLS")
		verifyTls(tcpAddr, u)
	}

}

func displayLoadingAnimation(duration time.Duration) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	done := time.After(duration)
	fmt.Print("Loading")

	for {
		select {
		case <-ticker.C:
			fmt.Print(".")
		case <-done:
			fmt.Print("\n")
			return
		}
	}
}

func displayResults(stats Statistics) {

	// yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	boldGreen1 := color.New(color.FgGreen).SprintFunc()
	// info := color.New(color.FgWhite, color.BgGreen).SprintFunc()
	fmt.Println("Results:")
	fmt.Printf(" Total Requests %s.......................: %s\n", boldGreen1("(2XX)"), boldGreen1(stats.TotalRequests))
	fmt.Printf(" Failed Requests %s......................: %s\n", red("(5XX)"), red(stats.FailedRequests))
	fmt.Printf(" Request/second.............................: %.2f\n", stats.RequestsPerSecond)

	fmt.Printf("\nTotal Request Time (s) (Min, Max, Mean).....: %.2f, %.2f, %.2f\n",
		stats.TotalRequestTime.Min.Seconds(), stats.TotalRequestTime.Max.Seconds(), stats.TotalRequestTime.Mean.Seconds())

	fmt.Printf("Time to First Byte (s) (Min, Max, Mean).....: %.2f, %.2f, %.2f\n",
		stats.TimeToFirstByte.Min.Seconds(), stats.TimeToFirstByte.Max.Seconds(), stats.TimeToFirstByte.Mean.Seconds())

	fmt.Printf("Time to Last Byte (s) (Min, Max, Mean)......: %.2f, %.2f, %.2f\n",
		stats.TimeToLastByte.Min.Seconds(), stats.TimeToLastByte.Max.Seconds(), stats.TimeToLastByte.Mean.Seconds())
}

func consuneUrlsfromFiles(filePath string) ([]string, error) {

	var urls []string

	file, err := os.Open(filePath)

	if err != nil {
		panic(err)

	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

//func to make request for tcp connection

// func makeRequest(tcpAddr *net.TCPAddr, u *url.URL, method string, contentType *string, content string, responseBody bool, headers bool) {
// 	// defer wg.Done()

// 	conn, err := net.DialTCP("tcp", nil, tcpAddr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// // result := net.DialTCP("tcp",u)
// 	defer conn.Close()

// 	// mu.Lock()
// 	// defer mu.Unlock()
// 	// fmt.Println(conn.RemoteAddr())
// 	request := fmt.Sprintf("%s %s HTTP/1.1\r\n", method, u.Path)
// 	request += fmt.Sprintf("Host: %s\r\n", u.Host)
// 	request += fmt.Sprintf("Content-Type: %s\r\n", *contentType)
// 	request += "Accept: */*\r\n"
// 	// fmt.Println(request)

// 	if method == "POST" || method == "PUT" {
// 		request += fmt.Sprintf("Content-Length: %d\r\n", len(content))
// 		request += fmt.Sprintf("Content-Type: %s\r\n", *contentType)
// 		request += "Connection: close\r\n\r\n"
// 		request += content
// 	} else {
// 		request += "Connection: close\r\n\r\n"
// 	}

// 	conn.Write([]byte(request))
// 	// fmt.Println(a, err)
// 	// fmt.Println("Request sent: ", request)
// 	var responseLines []string
// 	scanner := bufio.NewScanner(conn)

// 	for scanner.Scan() {
// 		// fmt.Println(scanner.Text())
// 		line := scanner.Text()
// 		if strings.HasPrefix(line, "HTTP/1.1") {
// 			if headers == false {
// 				fmt.Println("Response code", line)
// 				break
// 			}
// 		}
// 		if line == "" {
// 			break // Headers end
// 		}
// 		responseLines = append(responseLines, line)
// 	}

// 	if responseBody {
// 		for scanner.Scan() {
// 			line := scanner.Text()
// 			fmt.Println(line)
// 		}
// 	}

// }
