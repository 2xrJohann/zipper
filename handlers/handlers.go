package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (z *zipRequest) Parse(r *http.Request) error {
	if err := json.NewDecoder(r.Body).Decode(z); err != nil {
		return errors.New("failed to decode request body")
	}

	if len(z.Paths) == 0 {
		return errors.New(">=1 path is required")
	}

	return nil
}

func (z *zipRequestHandlers) ZipWithMutexHandler(w http.ResponseWriter, r *http.Request) {
	var zipReq zipRequest
	if badRequest := zipReq.Parse(r); badRequest != nil {
		http.Error(w, badRequest.Error(), http.StatusBadRequest)
		return
	}

	z.zipperService.WithMutex(zipReq.Paths)

	fmt.Fprintf(w, "Requested zipping with mutex: %v", zipReq.Paths)
}

func (z *zipRequestHandlers) ZipWithWriterChannel(w http.ResponseWriter, r *http.Request) {
	var zipReq zipRequest
	if badRequest := zipReq.Parse(r); badRequest != nil {
		http.Error(w, badRequest.Error(), http.StatusBadRequest)
		return
	}

	z.zipperService.WithWriterChannel(zipReq.Paths)

	fmt.Fprintf(w, "Requested zipping without mutex: %v", zipReq.Paths)
}
