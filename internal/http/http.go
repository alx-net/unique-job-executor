package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/alx-net/concurrent-task-executor/internal/uniquejob"
	"github.com/alx-net/concurrent-task-executor/internal/utils"
	"github.com/gorilla/mux"
)

type Response struct {
	Result string `json:"result"`
}

type JobExecutor struct {
	*uniquejob.JobExecutor[Response, uint64]
}

type WrapperFunc func(int64) Response

func routes(executor JobExecutor) {
	r := mux.NewRouter()

	r.HandleFunc("/fib/{num}", handleRequest(executor, fibonacciWrapper)).Methods("GET")
	r.HandleFunc("/isprime/{num}", handleRequest(executor, isPrimeWrapper)).Methods("GET")
	r.Use(validationMiddleware)

	log.Fatal(http.ListenAndServe(":8080", r))
}

func validate(vars map[string]string) error {
	num, err := strconv.ParseInt(vars["num"], 10, 64)

	// Validation
	if err != nil {
		return err
	} else if num < 0 {
		return fmt.Errorf("negative numbers are not allowed")
	}

	return nil
}

func validationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		validate(vars)
		next.ServeHTTP(w, r)
	})
}

func handleRequest(executor JobExecutor, next WrapperFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		num, _ := strconv.ParseInt(vars["num"], 10, 64)

		identifier := utils.HashFromStrings(r.RequestURI, vars["num"])

		res, err := handleJob(r.Context(), executor, identifier, num, next)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		err = encodeRequest(w, res)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
	}
}

func handleJob(
	ctx context.Context,
	executor JobExecutor,
	identifier uint64,
	num int64,
	next WrapperFunc) (Response, error) {

	job := uniquejob.NewJob(identifier,
		func(ctx context.Context) (Response, error) {
			return next(num), nil
		},
	)

	subscription := executor.Execute(ctx, job)

	return subscription.Subscribe(ctx)
}

func encodeRequest(w http.ResponseWriter, response Response) error {
	encResponse, err := json.Marshal(response)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(encResponse)
	return nil
}

func fibonacciWrapper(num int64) Response {
	return Response{Result: fmt.Sprint(utils.Fibonacci(num))}
}

func isPrimeWrapper(num int64) Response {
	return Response{Result: fmt.Sprint(utils.IsPrime(num))}
}
