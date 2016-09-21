package options

// Options defines methods that determine how clients connect with MongoDB
type Options interface {
	UseSSL() bool
	UseFIPSMode() bool
}
