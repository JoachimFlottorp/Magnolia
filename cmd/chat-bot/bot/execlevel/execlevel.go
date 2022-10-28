package execlevel

type ExecutionLevel int

const (
	ExecutionLevelEveryone = ExecutionLevel(iota)
	ExecutionLevelAdmin    = ExecutionLevel(iota)
)
