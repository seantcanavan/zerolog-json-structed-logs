package sl

import "time"

// ZLObjectKey is the key we use to map into the .Object() call to zerolog
const ZLObjectKey = "sl" // this is our global json property key for logged items

type ZLJSONItem struct {
	ErrorAsJSON map[string]any `json:"sl,omitempty"`
	Level       string         `json:"level,omitempty"`
	Message     string         `json:"message,omitempty"`
	Time        time.Time      `json:"time,omitempty"`
}
