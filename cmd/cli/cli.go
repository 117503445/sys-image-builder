package main

var cli struct {
	Build CmdBuild `cmd:"" help:"Build system image"`
}

type CmdBuild struct {
}

func (cmd *CmdBuild) Run() error {
	build()
	return nil
}
