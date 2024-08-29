package urlvalues // import "go.gideaworx.io/go-encoding/urlvalues"

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type urlValueTag struct {
	name       string
	omitEmpty  bool
	joinString string
}

func strSliceCheck(expectedValue string) func(string) bool {
	return func(s string) bool {
		return strings.HasPrefix(strings.TrimSpace(strings.ToLower(s)), strings.ToLower(expectedValue))
	}
}

func ParseTag(tag string) (*urlValueTag, error) {
	if tag == "-" {
		return nil, errSkip
	}

	parts := strings.Split(tag, ",")
	t := &urlValueTag{
		name: parts[0],
	}

	joinStartIndex := slices.IndexFunc(parts, strSliceCheck("join='"))
	joinEndIndex := -1
	if joinStartIndex > 0 {
		joinEndIndex = joinStartIndex + 1 + slices.IndexFunc(parts[joinStartIndex+1:], func(s string) bool {
			return strings.HasSuffix(s, "'")
		})
	}

	if joinEndIndex < joinStartIndex {
		return nil, errors.New(`tag had "join='..." but not a closing "'"`)
	}

	if joinStartIndex > 0 {
		switch joinEndIndex - joinStartIndex {
		case 0:
			t.joinString = strings.TrimSuffix(strings.TrimPrefix(parts[joinStartIndex], "join='"), "'")
		case 1:
			t.joinString = fmt.Sprintf("%s,%s",
				strings.TrimPrefix(parts[joinStartIndex], "join='"),
				strings.TrimSuffix(parts[joinEndIndex], "'"))
		default:
			subParts := make([]string, joinEndIndex-joinStartIndex+1)
			copy(subParts, parts[joinStartIndex:joinEndIndex+1])
			subParts[0] = strings.TrimPrefix(subParts[0], "join='")
			subParts[len(subParts)-1] = strings.TrimSuffix(subParts[len(subParts)-1], "'")
			t.joinString = strings.Join(subParts, ",")
		}
	}
	omitIndex := slices.IndexFunc(parts, strSliceCheck("omitempty"))
	if omitIndex < joinStartIndex || omitIndex > joinEndIndex {
		t.omitEmpty = (omitIndex > 0)
	}

	return t, nil
}
