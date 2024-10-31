package main

import (
	"github.com/alx-net/concurrent-task-executor/internal/http"
)

func main() {
	// start http server on port 8080 with min job duration of 5 seconds
	http.Routes(8080, 5)
}
