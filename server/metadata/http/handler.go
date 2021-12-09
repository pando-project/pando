package http

import (
	"Pando/internal/metrics"
	"Pando/statetree"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ipfs/go-cid"
	"net/http"
	"strconv"
)

// handler handles requests for the finder resource
type httpHandler struct {
	metaHandler *MetaHandler
}

type MetaHandler struct {
	StateTree *statetree.StateTree
}

func newHandler(stateTree *statetree.StateTree) *httpHandler {
	return &httpHandler{
		metaHandler: &MetaHandler{StateTree: stateTree},
	}
}

func (h *httpHandler) ListSnapShots(w http.ResponseWriter, r *http.Request) {
	record := metrics.APITimer(context.Background(), metrics.ListMetadataLatency)
	defer record()

	snapCidList, err := h.metaHandler.StateTree.GetSnapShotCidList()
	if err != nil {
		log.Error("cannot list snapshots, err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	resBytes, err := json.Marshal(snapCidList)
	if err != nil {
		log.Error("cannot list snapshots, err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, http.StatusOK, resBytes)
}

func (h *httpHandler) ListSnapShotInfo(w http.ResponseWriter, r *http.Request) {
	record := metrics.APITimer(context.Background(), metrics.ListSnapshotInfoLatency)
	defer record()

	cidStr, err := getSsCid(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ssCid, err := cid.Decode(cidStr)
	if err != nil {
		log.Error("cannot decode input cid, err", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	ss, err := h.metaHandler.StateTree.GetSnapShot(ssCid)
	if err != nil && err != statetree.NotFoundErr {
		log.Errorf("cannot get snapshot: %s, err: %s", ssCid.String(), err.Error())
		http.Error(w, "", http.StatusNotFound)
		return
	} else if err != nil {
		log.Errorf("not found snapshot: %s", ssCid.String())
		http.Error(w, "", http.StatusNotFound)
		return
	}

	resBytes, err := json.Marshal(ss)
	if err != nil {
		log.Error("cannot marshal snapshot, err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, http.StatusOK, resBytes)
}

func (h *httpHandler) GetSnapShotByHeight(w http.ResponseWriter, r *http.Request) {
	record := metrics.APITimer(context.Background(), metrics.GetSnapshotByHeightLatency)
	defer record()

	ssHeight, err := getSsHeight(r)
	if err != nil {
		log.Warnf("error input %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ss, err := h.metaHandler.StateTree.GetSnapShotByHeight(ssHeight)
	if err != nil {
		log.Warnf("cannot get snapshot by height: %d, err: %s", ssHeight, err.Error())
		http.Error(w, "", http.StatusNotFound)
		return
	}

	resBytes, err := json.Marshal(ss)
	if err != nil {
		log.Error("cannot marshal snapshot, err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	WriteJsonResponse(w, http.StatusOK, resBytes)
}

func getSsCid(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	id := vars["sscid"]
	if id == "" {
		return "", fmt.Errorf("invalid cid to search")
	}
	return id, nil
}

func getSsHeight(r *http.Request) (uint64, error) {
	vars := mux.Vars(r)
	id := vars["height"]
	if id == "" {
		return uint64(0), fmt.Errorf("invalid height to search")
	}
	height, err := strconv.Atoi(id)
	if err != nil {
		return uint64(0), fmt.Errorf("invalid height to search, %s", id)
	}
	return uint64(height), nil
}

func WriteJsonResponse(w http.ResponseWriter, status int, body []byte) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if _, err := w.Write(body); err != nil {
		log.Errorw("cannot write response", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}
