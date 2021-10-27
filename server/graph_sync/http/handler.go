package http

import (
	"Pando/legs"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ipfs/go-datastore"
	"net/http"
)

// handler handles requests for the finder resource
type httpHandler struct {
	graphSyncHandler *GraphSyncHandler
}

type GraphSyncHandler struct {
	Core *legs.LegsCore
}

func newHandler(core *legs.LegsCore) *httpHandler {
	return &httpHandler{
		graphSyncHandler: &GraphSyncHandler{core},
	}
}

func (h *httpHandler) SubTopic(w http.ResponseWriter, r *http.Request) {
	topic, err := getTopic(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = h.graphSyncHandler.Core.NewSubscriber(topic)
	if err != nil {
		log.Error("cannot create subscriber", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *httpHandler) GetGraph(w http.ResponseWriter, r *http.Request) {
	id, err := getCid(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	v, err := h.graphSyncHandler.Core.DS.Get(datastore.NewKey(id))
	if err != nil {
		log.Error("cannot search for cid: ", id, "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	WriteJsonResponse(w, http.StatusOK, v)

	//topic, err := getTopic(r)
	//if err != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}
	//_, err = h.graphSyncHandler.Core.NewSubscriber(topic)
	//if err != nil {
	//	log.Error("cannot create subscriber", "err", err)
	//	http.Error(w, "", http.StatusInternalServerError)
	//	return
	//}
	//w.WriteHeader(http.StatusOK)
}

func getTopic(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	topic := vars["topic"]
	if topic == "" {
		return "", fmt.Errorf("invalid topic to subscribe")
	}
	return topic, nil
}

func getCid(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		return "", fmt.Errorf("invalid cid to search")
	}
	return id, nil
}

func WriteJsonResponse(w http.ResponseWriter, status int, body []byte) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if _, err := w.Write(body); err != nil {
		log.Errorw("cannot write response", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}
