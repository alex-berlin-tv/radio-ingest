package stackfield

// A Stackfield room to which messages can be sent.
type Room struct {
	Url string
}

// Returns a new instance of [Room].
func NewRoom(url string) Room {
	return Room{
		Url: url,
	}
}
