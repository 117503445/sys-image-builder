package main

var cli struct {
	Build CmdBuild `cmd:"" help:"Build system image"`
	Push  CmdPush  `cmd:"" help:"Push system image to registry"`
}

type CmdBuild struct {
}

func (cmd *CmdBuild) Run() error {
	build()
	return nil
}

type CmdPush struct {
	Image    string `help:"Image reference (e.g. registry:5000/repo:tag)" required:""`
	Username string `help:"Registry username" env:"REGISTRY_USERNAME"`
	Password string `help:"Registry password" env:"REGISTRY_PASSWORD"`
	Insecure bool   `help:"Use HTTP instead of HTTPS" default:"false"`
}

func (cmd *CmdPush) Run() error {
	return push(cmd)
}
