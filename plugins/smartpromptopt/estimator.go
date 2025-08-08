package smartpromptopt

import (
	"unicode/utf8"

	"github.com/maximhq/bifrost/core/schemas"
)

type TokenEstimator interface {
	Estimate(req *schemas.BifrostRequest) int
}

type DefaultEstimator struct{}

func (DefaultEstimator) Estimate(req *schemas.BifrostRequest) int {
	if req == nil { return 0 }
	chars := 0
	if req.Input.TextCompletionInput != nil {
		chars += utf8.RuneCountInString(*req.Input.TextCompletionInput)
	}
	if req.Input.ChatCompletionInput != nil {
		for _, m := range *req.Input.ChatCompletionInput {
			chars += utf8.RuneCountInString(m.Content)
		}
	}
	// crude estimate: 1 token ~ 4 chars
	if chars <= 0 { return 0 }
	return (chars + 3) / 4
}
