package notification

type None struct{}

func NewNone() *None {
	return &None{}
}

func (n *None) Notify(message string) {
}
