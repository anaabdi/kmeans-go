package controller

import (
	"net/http"
)

func ListDataController(rw http.ResponseWriter, r *http.Request) {
	if len(mainNodes) == 0 {
		Write(rw, http.StatusOK, map[string][]Node{"datasets": []Node{}})
		return
	}
	Write(rw, http.StatusOK, map[string][]Node{"datasets": mainNodes})
}
