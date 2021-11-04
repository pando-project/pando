package http

import (
	"Pando/legs"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/peer"
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

func (h *httpHandler) SubProvider(w http.ResponseWriter, r *http.Request) {
	peerID, err := getProviderID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.graphSyncHandler.Core.Subscribe(context.Background(), peerID)
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

func getProviderID(r *http.Request) (peer.ID, error) {
	vars := mux.Vars(r)
	pid := vars["peerid"]
	providerID, err := peer.Decode(pid)
	if err != nil {
		return providerID, fmt.Errorf("cannot decode provider id: %s", err)
	}
	return providerID, nil
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
