package notification

// WebexMessageRequest is the Webex Teams Create Message Request Parameters
type WebexMessageRequest struct {
	RoomID        string   `json:"roomId,omitempty"`        // Room ID.
	ToPersonID    string   `json:"toPersonId,omitempty"`    // Person ID (for type=direct).
	ToPersonEmail string   `json:"toPersonEmail,omitempty"` // Person email (for type=direct).
	Text          string   `json:"text,omitempty"`          // Message in plain text format.
	Markdown      string   `json:"markdown,omitempty"`      // Message in markdown format.
	Files         []string `json:"files,omitempty"`         // File URL array.
}
