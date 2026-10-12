package connect

import (
	"encoding/json"

	"github.com/charmbracelet/x/ansi"
)

// Plain is the inputs of a page as the designer should draw them: every text without the
// colour codes the app puts in some of its texts, since the design owns the colours. Rows
// and lists are walked too. A value that cannot be written as JSON is left as it was.
func Plain(inputs map[string]any) map[string]any {
	raw, err := json.Marshal(inputs)
	if err != nil {
		return inputs
	}
	var out map[string]any
	if json.Unmarshal(raw, &out) != nil {
		return inputs
	}
	for k, v := range out {
		out[k] = stripped(v)
	}
	return out
}

func stripped(v any) any {
	switch x := v.(type) {
	case string:
		return ansi.Strip(x)
	case []any:
		for i := range x {
			x[i] = stripped(x[i])
		}
	case map[string]any:
		for k := range x {
			x[k] = stripped(x[k])
		}
	}
	return v
}
