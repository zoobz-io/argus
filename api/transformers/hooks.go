package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// HookToResponse transforms a Hook model to an API response (omits secret).
func HookToResponse(h *models.Hook) wire.HookResponse {
	return wire.HookResponse{
		ID:        h.ID,
		URL:       h.URL,
		Active:    h.Active,
		CreatedAt: h.CreatedAt,
	}
}

// HookToCreateResponse transforms a Hook model to a create response (includes secret).
func HookToCreateResponse(h *models.Hook) wire.HookCreateResponse {
	return wire.HookCreateResponse{
		ID:        h.ID,
		URL:       h.URL,
		Secret:    h.Secret,
		Active:    h.Active,
		CreatedAt: h.CreatedAt,
	}
}

// HooksToResponse transforms a slice of Hook models to responses.
func HooksToResponse(hooks []*models.Hook) []wire.HookResponse {
	result := make([]wire.HookResponse, len(hooks))
	for i, h := range hooks {
		result[i] = HookToResponse(h)
	}
	return result
}

// DeliveryToResponse transforms a Delivery model to an API response.
func DeliveryToResponse(d *models.Delivery) wire.DeliveryResponse {
	resp := wire.DeliveryResponse{
		ID:         d.ID,
		HookID:     d.HookID,
		EventID:    d.EventID,
		StatusCode: d.StatusCode,
		Attempt:    d.Attempt,
		CreatedAt:  d.CreatedAt,
	}
	if d.Error != nil {
		resp.Error = *d.Error
	}
	return resp
}

// DeliveriesToListResponse transforms an OffsetResult of deliveries to a list response.
func DeliveriesToListResponse(result *models.OffsetResult[models.Delivery]) wire.DeliveryListResponse {
	deliveries := make([]wire.DeliveryResponse, len(result.Items))
	for i, d := range result.Items {
		deliveries[i] = DeliveryToResponse(d)
	}
	return wire.DeliveryListResponse{
		Deliveries: deliveries,
		Offset:     result.Offset,
		Limit:      len(result.Items),
		Total:      result.Total,
	}
}
