package contact

// Message - the model for data sent over a Connection
type Message struct {
	Meta interface{} `json:"meta"`
	Body string      `json:"body"`
}

// IsEmpty - return a bool indicating whether a Message is empty
func (m *Message) IsEmpty() bool {
	return m.Meta == nil && m.Body == ""
}
