package bot

type Bot interface {
	Run() error
	Stop() error
	Join(string)
	Say(string, string, ...interface{})
}
