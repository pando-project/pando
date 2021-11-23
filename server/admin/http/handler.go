package httpadminserver

import (
	"Pando/internal/handler"
	"Pando/internal/httpserver"
	"Pando/internal/registry"
	"io"
	"net/http"
)

type httpHandler struct {
	adminHandler *handler.AdminHandler
}

func newHandler(registry *registry.Registry) *httpHandler {
	return &httpHandler{
		adminHandler: handler.NewAdminHandler(registry),
	}
}

// POST /providers
func (h *httpHandler) RegisterProvider(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorw("failed reading body", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.adminHandler.RegisterProvider(body)
	if err != nil {
		httpserver.HandleError(w, err, "register")
		log.Warnf("register failed: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}
