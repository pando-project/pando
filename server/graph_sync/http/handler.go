package http

import (
	"Pando/legs"
	"fmt"
	"github.com/gorilla/mux"
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

func getTopic(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	topic := vars["topic"]
	if topic == "" {
		return "", fmt.Errorf("invalid topic to subscribe")
	}
	return topic, nil
}
