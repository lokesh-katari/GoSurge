package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"

	// "reflect"
	"strings"
	"sync"
	"time"
	// "strings"
)

var wg sync.WaitGroup

var mu sync.Mutex

// https://www.google.com/
func makeRequest(tcpAddr *net.TCPAddr, u *url.URL, method string, contentType *string, content string, responseBody bool, headers bool) {
	// defer wg.Done()

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		panic(err)
	}
	// // result := net.DialTCP("tcp",u)
	defer conn.Close()

	// mu.Lock()
	// defer mu.Unlock()
	// fmt.Println(conn.RemoteAddr())
	request := fmt.Sprintf("%s %s HTTP/1.1\r\n", method, u.Path)
	request += fmt.Sprintf("Host: %s\r\n", u.Host)
	request += fmt.Sprintf("Content-Type: %s\r\n", *contentType)
	request += "Accept: */*\r\n"
	// fmt.Println(request)

	if method == "POST" || method == "PUT" {
		request += fmt.Sprintf("Content-Length: %d\r\n", len(content))
		request += fmt.Sprintf("Content-Type: %s\r\n", *contentType)
		request += "Connection: close\r\n\r\n"
		request += content
	} else {
		request += "Connection: close\r\n\r\n"
	}

	conn.Write([]byte(request))
	// fmt.Println(a, err)
	// fmt.Println("Request sent: ", request)
	var responseLines []string
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		line := scanner.Text()
		if strings.HasPrefix(line, "HTTP/1.1") {
			if headers == false {
				fmt.Println("Response code", line)
				break
			}
		}
		if line == "" {
			break // Headers end
		}
		responseLines = append(responseLines, line)
	}

	if responseBody {
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
		}
	}

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

	if len(os.Args) < 2 {
		fmt.Println("Usage: mycurl <URL>")
		os.Exit(1)
	}
	method := flag.String("m", "GET", "`method`")
	urlString := flag.String("url", "", "url")
	contentType := flag.String("con-type", "text/plain", "Content type")
	help := flag.Bool(("help"), false, "help")
	content := flag.String("content", "", "Content for POST or PUT")
	responseBody := flag.Bool("res-body", false, "Response body")
	reqNum := flag.Int("n", 1, "Number of requests")
	conqreq := flag.Int("c", 1, "Number of concurrent requests")
	verTls := flag.Bool("tls", false, "Verify TLS")
	headers := flag.Bool("H", false, "Headers")
	flag.Parse()
	// fmt.Printf("this %s\n", method)
	// fmt.Println(*method, *urlString)
	if *urlString == "" {
		fmt.Println("Error: Server address is required.")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *help {
		fmt.Println("Usage: mycurl <URL>")
		flag.PrintDefaults()
	}
	u, err := url.Parse(*urlString)

	// port := 443
	// fmt.Println(u, err)
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "80" // Default port for HTTP
	}
	// fmt.Println(host)
	serverAddr := fmt.Sprintf("%s:%s", host, port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	// typ := reflect.TypeOf(u)
	// fmt.Printf("Type of %v: %v\n", tcpAddr, typ)
	if err != nil {
		panic(err)
	}

	// fmt.Println(*contentType, *content, *responseBody, *reqNum)

	conqreqChan := make(chan struct{}, *conqreq)
	for i := 0; i < *reqNum; i++ {
		wg.Add(1)
		conqreqChan <- struct{}{}

		go func() {
			defer wg.Done()
			start := time.Now()
			makeRequest(tcpAddr, u, *method, contentType, *content, *responseBody, *headers)
			fmt.Printf("Request took: %v\n", time.Since(start))
			<-conqreqChan
		}()
	}
	// Wait for all requests to finish
	wg.Wait()
	if *verTls {
		fmt.Println("\n\n\rVerifying TLS")
		verifyTls(tcpAddr, u)
	}

}
