package resharing

import (
	"github.com/Datura-ai/relayer/packages/service/tss/blame"
	"github.com/Datura-ai/relayer/packages/service/tss/common"
)

// Response keygen response
type Response struct {
	Status common.Status `json:"status"`
	Blame  blame.Blame   `json:"blame"`
}

// NewResponse create a new instance of keygen.Response
func NewResponse(status common.Status, blame blame.Blame) Response {
	return Response{
		Status: status,
		Blame:  blame,
	}
}
