# GoSurge

GoSurge is a simple HTTP benchmarking and load testing tool written in Go. It takes advantage of Go's concurrency, supports various HTTP methods (GET, POST, PUT, DELETE), and can read URLs from files for making concurrent requests.

## Features

- **Concurrency:** Leverages Go's concurrency for efficient and parallelized HTTP requests.
- **HTTP Methods:** Supports popular HTTP methods like GET, POST, PUT, and DELETE.
- **URLs from Files:** Read URLs from files and make concurrent requests accordingly.
- **TLS Configuration Testing:** Verify TLS configurations for specified links.
- **Combination of Curl and Load Testing:** Provides functionality for both making HTTP requests and load testing.


# Getting Started
## Installation
### Clone the repository:
 ```bash
    
  git clone https://github.com/lokesh-katari/GoSurge.git

    
 ```
### Navigate to the project directory:

 ```bash
    
    cd GoSurge
    
 ```
Build the executable:
### Linux

1. Build the executable:

 ```bash
    
   go build -o gosurge gosurge.go
    
 ```

   This will create an executable named gosurge in the same directory.

## Move to Bin Directory:

To make it system-wide, move the gosurge executable to the /usr/local/bin directory. This directory is usually in the system's PATH.

 ```bash
    
    sudo mv gosurge /usr/local/bin/
    
 ```
## Usage:

Now you can use gosurge from any directory in the terminal.
### For Windows:
 
 Open a command prompt in the directory containing gosurge.go.

 Run the following command to build the executable:

Copy code
 ```bash
    
    go build -o gosurge.exe gosurge.go
    
 ```
This will create an executable named gosurge.exe in the same directory.

Add to System PATH:

Add the directory containing gosurge.exe to your system's PATH. You can do this through the system settings.
Usage:

Now you can use gosurge from any directory in the command prompt.
Usage
```bash

gosurge -m GET -url http://example.com -n 10 -c 5

```
```bash
-m: HTTP method (GET, POST, PUT, DELETE).
-url: Target URL.
-n: Number of requests to make.
-c: Number of concurrent requests.
-res-body :If you want the response body
-tls=true :if you want to get tls configuration
For more options, run gosurge --help.
```
Contributing
Feel free to contribute! Start by forking the repository and opening a pull request. Your contributions are appreciated.


# DEMO

https://github.com/lokesh-katari/GoSurge/assets/111894942/2bc9cab7-5954-419d-984e-07bb2b824036



