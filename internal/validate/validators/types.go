package validators

import (
	"github.com/nyambati/fuse/internal/diag"
)

type Validator interface {
	Validate() []diag.Diagnostic
}
