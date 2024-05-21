package engine

// SampleEvent is a struct that represents an event with relevant fields
type SampleEvent struct {
	Image       string
	CommandLine string
	ParentImage string
}

// Keywords implements the Keyworder interface
func (e SampleEvent) Keywords() ([]string, bool) {
	return []string{e.CommandLine}, true
}

// Select implements the Selector interface
func (e SampleEvent) Select(field string) (interface{}, bool) {
	switch field {
	case "Image":
		return e.Image, true
	case "CommandLine":
		return e.CommandLine, true
	case "ParentImage":
		return e.ParentImage, true
	default:
		return nil, false
	}
}
