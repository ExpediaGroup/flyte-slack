// types is for storing shared data structures
package types

// Conversation describes slack channel
type Conversation struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Topic string `json:"topic"`
}

type Reactions struct {
}
