package main

import (
	"flag"
)

var (
	source = flag.String("src", "http://localhost:9411/api/v2", "URL of zipkin API")
)

func init() {
	flag.Parse()
}

func main() {

}
