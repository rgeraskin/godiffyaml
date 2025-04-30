package docs

import (
	"fmt"
	"strings"
)

type Doc map[string]any

// GetValueByPath returns the value at the given dot-separated path
func (d *Doc) GetValueByPath(path string) string {
	current := *d
	path = strings.TrimPrefix(path, ".")
	parts := strings.Split(path, ".")

	for i, part := range parts {
		if current == nil {
			return ""
		}

		if i == len(parts)-1 {
			// Last part - guess the type
			switch v := current[part].(type) {
			case string:
				return v
			case int:
				return fmt.Sprintf("%d", v)
			case float64:
				return fmt.Sprintf("%g", v)
			case bool:
				return fmt.Sprintf("%v", v)
			default:
				return ""
			}
		}

		// Not the last part - try to go deeper
		if next, ok := current[part].(Doc); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}

type Docs struct {
	Docs  []Doc
	Order []string // Sorting criteria paths
}

func (d Docs) Len() int      { return len(d.Docs) }
func (d Docs) Swap(i, j int) { d.Docs[i], d.Docs[j] = d.Docs[j], d.Docs[i] }
func (d Docs) Less(i, j int) bool {
	for _, path := range d.Order {
		valI := d.Docs[i].GetValueByPath(path)
		valJ := d.Docs[j].GetValueByPath(path)
		if valI != valJ {
			return valI < valJ
		}
	}
	return false
}
