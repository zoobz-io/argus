// Package vocabulary provides a pipeline for validating and classifying
// vocabulary entries (topics and tags) before persistence.
package vocabulary

// Context carries data through the vocabulary pipeline stages.
type Context struct {
	TenantID    string
	Name        string
	Description string
	ID          string // Set on update, empty on create.
}

// Clone returns a copy of the context.
func (vc *Context) Clone() *Context {
	c := *vc
	return &c
}
