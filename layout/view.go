package layout

type View interface {
	Draw(context DrawingContext) error
	GetName() string
	GetKeyBindings() []*KeyBinding
}
