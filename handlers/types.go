package handlers

import (
	"net/http"
	"runtime"
	"threadpoolcompress/service"
)

const MAX_ROUTINES = 100

type zipRequestHandlers struct {
	zipperService service.ZipperService
}

func New(zipperService service.ZipperService) *zipRequestHandlers {
	return &zipRequestHandlers{
		zipperService: zipperService,
	}
}

type zipRequest struct {
	Paths []string `json:"paths"`
}

type MiddlewareHandler func(http.HandlerFunc) http.HandlerFunc

func (z *zipRequestHandlers) HttpMethod(expectedMethod string) MiddlewareHandler {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != expectedMethod {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			handler(w, r)
		}
	}
}

func (z *zipRequestHandlers) WithSharedGoRoutineLimit(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		routineCount := runtime.NumGoroutine()

		if routineCount >= MAX_ROUTINES {
			http.Error(w, "Working too hard, try again later", http.StatusServiceUnavailable)
			return
		}
		handler(w, r)
	}
}

func (z *zipRequestHandlers) ApplyMiddlewares(handler http.HandlerFunc, middlewares ...MiddlewareHandler) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}
