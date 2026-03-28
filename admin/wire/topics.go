package wire

import "time"

// AdminTopicResponse is the wire representation of a single topic.
type AdminTopicResponse struct {
	CreatedAt   time.Time `json:"created_at" description:"Creation timestamp"`
	UpdatedAt   time.Time `json:"updated_at" description:"Last update timestamp"`
	Name        string    `json:"name" description:"Topic name" example:"Security"`
	Description string    `json:"description" description:"Topic description" example:"Security-related documents"`
	ID          string    `json:"id" description:"Topic ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID    string    `json:"tenant_id" description:"Owning tenant ID" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// Clone returns a shallow copy of the response.
func (t AdminTopicResponse) Clone() AdminTopicResponse {
	return t
}

// AdminTopicListResponse is the wire representation of a topic list.
type AdminTopicListResponse struct {
	Topics []AdminTopicResponse `json:"topics" description:"List of topics"`
}

// Clone returns a deep copy of the response.
func (r AdminTopicListResponse) Clone() AdminTopicListResponse {
	c := r
	if r.Topics != nil {
		c.Topics = make([]AdminTopicResponse, len(r.Topics))
		copy(c.Topics, r.Topics)
	}
	return c
}
