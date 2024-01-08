# MASII Documentation

## Installation

1. Unzip the folder.
2. Navigate to `/binaries`
2. Select the binary based off your operating system and platform, eg `masii-linux-amd64`  for Linux operating systems that use AMD64 architecture.
3. `chmod +x masii-linux-amd64`

It's important to note that not every architecture has been statically compiled. If the binary needed cannot be found. Install Go 1.13.8 and execute `go build main.go`. 

## Usage

Prompt the help page via the following:
```go
./masii-linux-amd64 -h
```

Below are some example usages:

Passing two Cookies, for two seperate user roles. A rate-limit of max 10 requests sent per second with the crawling output enabled and running the Headers security module.
```go
./masii-linux-amd64 -u http://127.0.0.1:5000 -a "Cookie: session=eyJ1c2VyX2lkIjoxfQ.Y9lP9g.stwoa_Vlxa_xqajmvx_gVQBjujw, session=eyJ1c2VyX2lkIjoyfQ.ZAdbrQ.BK5tuLMylUpr1jZY1Ztv8TR2RhE" -s -t Admin,JOE -r 10 -c -m headers
```

