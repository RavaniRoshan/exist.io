package origin

import (
	"net/http"
	"time"

	"exist.io/internal/auth"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
	"exist.io/pkg/scopes"
)

var (
	reqWebhookDeliveriesRead = scopes.Requirements{
		TokenTypes: []string{scopes.TypeApp, scopes.TypeUser},
		Scopes:     []string{scopes.AppWebhookDeliveriesRead},
		Ambient:    true,
	}
	reqWebhookDeliveriesWrite = scopes.Requirements{
		TokenTypes: []string{scopes.TypeApp, scopes.TypeUser},
		Scopes:     []string{scopes.AppWebhookDeliveriesWrite},
		Ambient:    true,
	}
)

func (s *Server) registerWebhooks(mux *http.ServeMux) {
	s.register("OriginService_ListWebhookDeliveries", reqWebhookDeliveriesRead)
	s.register("OriginService_BatchRedeliverWebhookDeliveries", reqWebhookDeliveriesWrite)
	s.register("OriginService_PingWebhook", reqWebhookDeliveriesWrite)

	mux.Handle("GET /v1/origin/app/webhook/deliveries",
		s.guard("OriginService_ListWebhookDeliveries", s.listWebhookDeliveries))
	mux.Handle("POST /v1/origin/app/webhook/deliveries:batchRedeliver",
		s.guard("OriginService_BatchRedeliverWebhookDeliveries", s.batchRedeliverWebhookDeliveries))
	mux.Handle("POST /v1/origin/app/webhook/pings",
		s.guard("OriginService_PingWebhook", s.pingWebhook))
}

func (s *Server) getActingAppID(r *http.Request) string {
	if tok, ok := auth.ContextToken(r.Context()); ok && tok.Type() == scopes.TypeApp {
		return tok.Subject()
	}
	return "app_default"
}

func (s *Server) listWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	appID := s.getActingAppID(r)

	deliveries, err := s.store.ListWebhookDeliveries(r.Context(), appID, 100, 0)
	if err != nil {
		internalError(w, "list webhook deliveries: "+err.Error())
		return
	}

	var dtos []map[string]any
	for _, d := range deliveries {
		dtos = append(dtos, toWebhookDeliveryDTO(d))
	}
	if dtos == nil {
		dtos = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"deliveries":    dtos,
		"nextPageToken": "",
	})
}

type batchRedeliverReq struct {
	DeliveryIDs []string `json:"deliveryIds"`
}

func (s *Server) batchRedeliverWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	var req batchRedeliverReq
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}
	if len(req.DeliveryIDs) == 0 || len(req.DeliveryIDs) > 100 {
		badRequest(w, "deliveryIds must have between 1 and 100 entries")
		return
	}

	var results []map[string]any
	for _, id := range req.DeliveryIDs {
		results = append(results, map[string]any{
			"deliveryId": id,
			"status":     "scheduled",
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"results": results,
	})
}

func (s *Server) pingWebhook(w http.ResponseWriter, r *http.Request) {
	appID := s.getActingAppID(r)

	deliveryID := idgen.NewWebhookDelivery()
	eventID := idgen.NewEvent()
	now := time.Now().UTC().Format(time.RFC3339)

	del := store.WebhookDelivery{
		ID:            deliveryID,
		AppID:         appID,
		Event:         "ping",
		Payload:       `{"zen":"Exist smoothly."}`,
		Signature:     "v1ed_test_signature",
		Delivered:     true,
		DeliveredAt:   now,
		LastAttemptAt: now,
		AttemptCount:  1,
		CreatedAt:     now,
	}
	_, _ = s.store.CreateWebhookDelivery(r.Context(), del)

	writeJSON(w, http.StatusOK, map[string]any{
		"deliveryId":         deliveryID,
		"eventId":            eventID,
		"delivered":          true,
		"responseStatusCode": 200,
	})
}

func toWebhookDeliveryDTO(d store.WebhookDelivery) map[string]any {
	dto := map[string]any{
		"id": d.ID,
		"event": map[string]any{
			"id":   "evt_" + d.ID,
			"type": d.Event,
		},
		"createdAt": d.CreatedAt,
	}
	if d.InstallationID != "" {
		dto["installation"] = map[string]any{
			"id": d.InstallationID,
		}
	}
	if d.Delivered {
		dto["deliveredAt"] = d.DeliveredAt
		dto["lastAttempt"] = map[string]any{
			"responseStatusCode": 200,
			"status":             "succeeded",
			"attemptedAt":        d.LastAttemptAt,
		}
	}
	return dto
}
