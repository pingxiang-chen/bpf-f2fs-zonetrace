package server

import "net/http"

type Serializable interface {
	Serialize() []byte
}

func WriteJsonResponse(w http.ResponseWriter, response Serializable) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(response.Serialize())
	if err != nil {
		http.Error(w, "Error writing data", http.StatusInternalServerError)
		return
	}
	return
}

func WriteProtoBuf(w http.ResponseWriter, response Serializable) {
	w.Header().Set("Content-Type", "application/protobuf")
	_, err := w.Write(response.Serialize())
	if err != nil {
		http.Error(w, "Error writing data", http.StatusInternalServerError)
		return
	}
	return
}
