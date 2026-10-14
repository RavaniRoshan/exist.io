package origin

import (
	"encoding/json"
	"net/http"
	"testing"

	"exist.io/internal/store"
)

func TestWebhooks_Handlers(t *testing.T) {
	_, db, h := testServer(t)

	// Seed app_default for foreign key requirement
	_ = db.CreateNamespace(t.Context(), store.Namespace{ID: "ns_default", Slug: "default", Type: "team", Name: "Default"})
	_, _ = db.CreateApp(t.Context(), store.App{ID: "app_default", NamespaceID: "ns_default", DisplayName: "Default App"})

	// 1. PingWebhook
	res := do(h, "POST", "/v1/origin/app/webhook/pings", "{}")
	if res.Code != http.StatusOK {
		t.Fatalf("ping webhook expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var pingResp struct {
		DeliveryID         string `json:"deliveryId"`
		EventID            string `json:"eventId"`
		Delivered          bool   `json:"delivered"`
		ResponseStatusCode int    `json:"responseStatusCode"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &pingResp)
	if pingResp.DeliveryID == "" || pingResp.EventID == "" || !pingResp.Delivered || pingResp.ResponseStatusCode != 200 {
		t.Fatalf("unexpected ping response: %+v", pingResp)
	}

	// 2. ListWebhookDeliveries
	res = do(h, "GET", "/v1/origin/app/webhook/deliveries", "")
	if res.Code != http.StatusOK {
		t.Fatalf("list deliveries expected 200, got %d", res.Code)
	}
	var listResp struct {
		Deliveries []map[string]any `json:"deliveries"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &listResp)
	if len(listResp.Deliveries) != 1 {
		t.Fatalf("expected 1 delivery from ping, got %d", len(listResp.Deliveries))
	}

	// 3. BatchRedeliverWebhookDeliveries
	redeliverBody := `{"deliveryIds":["` + pingResp.DeliveryID + `"]}`
	res = do(h, "POST", "/v1/origin/app/webhook/deliveries:batchRedeliver", redeliverBody)
	if res.Code != http.StatusOK {
		t.Fatalf("batch redeliver expected 200, got %d: %s", res.Code, res.Body.String())
	}
	var redeliverResp struct {
		Results []map[string]any `json:"results"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &redeliverResp)
	if len(redeliverResp.Results) != 1 || redeliverResp.Results[0]["status"] != "scheduled" {
		t.Fatalf("unexpected redeliver response: %+v", redeliverResp)
	}
}
