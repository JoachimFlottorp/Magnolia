package bot

type Bot interface {
	Run() error
	Say(string, string, ...interface{})
}
