package rot

type EventType uint64

type Event struct {
	Type EventType
	Data map[string]any
}
