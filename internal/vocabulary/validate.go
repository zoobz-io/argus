package vocabulary

import (
	"context"
	"fmt"

	"github.com/zoobz-io/check"
	"github.com/zoobz-io/pipz"

	"github.com/zoobz-io/argus/internal/validate"
)

// ValidateID is the stage identity for vocabulary validation.
var ValidateID = pipz.NewIdentity("vocabulary-validate", "Validate vocabulary name and description")

func newValidateStage() pipz.Chainable[*Context] {
	return pipz.Apply(ValidateID, func(_ context.Context, vc *Context) (*Context, error) {
		err := check.All(
			validate.Name(vc.Name, "name"),
			validate.Description(vc.Description, "description"),
		).Err()
		if err != nil {
			return vc, fmt.Errorf("vocabulary validation failed: %w", err)
		}
		return vc, nil
	})
}
