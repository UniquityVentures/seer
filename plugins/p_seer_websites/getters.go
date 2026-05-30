package p_seer_websites

import (
	"context"
	"fmt"
	"strings"

	"github.com/UniquityVentures/lamu/getters"
	"github.com/UniquityVentures/lamu/lamu"
)

// pageURLStringFromKey reads a flattened struct URL under structMapKey (e.g. "$in.URL" or
// "$row.URL"). Value is [lamu.PageURL], string, or other (fmt.Sprint).
func pageURLStringFromKey(structMapKey string) getters.Getter[string] {
	return getters.Map(getters.Key[any](structMapKey), func(_ context.Context, v any) (string, error) {
		switch t := v.(type) {
		case nil:
			return "", nil
		case string:
			return strings.TrimSpace(t), nil
		case lamu.PageURL:
			return t.String(), nil
		default:
			return strings.TrimSpace(fmt.Sprint(t)), nil
		}
	})
}
