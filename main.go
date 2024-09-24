package main

import (
	"github.com/alx-net/concurrent-task-executor/internal/http"
)

func main() {
	http.Routes(8080, 5)
}
