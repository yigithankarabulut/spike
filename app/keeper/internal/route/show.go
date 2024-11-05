//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spiffe/spike/app/keeper/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

func routeShow(r *http.Request, w http.ResponseWriter) {
	fmt.Println("routeShow:", r.Method, r.URL.Path, r.URL.RawQuery)

	body := net.ReadRequestBody(r, w)
	if body == nil {
		return
	}

	var req reqres.RootKeyReadRequest
	if err := net.HandleRequestError(w, json.Unmarshal(body, &req)); err != nil {
		log.Println("routeShow: Problem handling request:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Println("routeShow: Problem writing response:", err.Error())
		}
		return
	}

	rootKey := state.RootKey()

	res := reqres.RootKeyReadResponse{RootKey: rootKey}
	md, err := json.Marshal(res)
	if err != nil {
		log.Println("routeShow: Problem generating response:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, err = io.WriteString(w, string(md))
	if err != nil {
		log.Println("routeShow: Problem writing response:", err.Error())
	}
}
