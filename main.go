package main

import (
	"fmt"
	"log"
	"net/http"
	"threadpoolcompress/handlers"
	"threadpoolcompress/service"
)

const PORT = ":8080"

func main() {
	zipperService := service.NewZipper()
	h := handlers.New(zipperService)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/zip/with-mutex",
		h.ApplyMiddlewares(
			h.ZipWithMutexHandler,
			h.HttpMethod(http.MethodPost),
			h.WithSharedGoRoutineLimit,
		),
	)

	mux.HandleFunc(
		"/zip/with-writer-channel",
		h.ApplyMiddlewares(
			h.ZipWithWriterChannel,
			h.HttpMethod(http.MethodPost),
			h.WithSharedGoRoutineLimit,
		),
	)

	fmt.Printf("Server starting on port %s...\n", PORT)
	log.Fatal(http.ListenAndServe(PORT, mux))
}
