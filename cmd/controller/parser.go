package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const (
	jsonContentType = "application/json"
	multipartType   = "multipart/data"
	jsonCharset     = "utf-8"
)

func ParseJSON(r *http.Request, body interface{}) error {
	return json.NewDecoder(r.Body).Decode(&body)
}

func Write(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(status)
	if data == nil {
		return
	}

	content, err := json.Marshal(data)
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	_, _ = w.Write(content)
}

type Response struct {
	Code   string      `json:"code"`
	Msg    string      `json:"message"`
	Detail string      `json:"detail,omitempty"`
	Data   interface{} `json:"data"`
}

func Error(w http.ResponseWriter, err error) {
	Write(w, http.StatusInternalServerError, Response{
		Code:   "4004",
		Msg:    "We failed to process your request",
		Detail: err.Error(),
	})
}
