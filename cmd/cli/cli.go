package main

var cli struct {
	Build CmdBuild `cmd:"" help:"Build system image"`
	Push  CmdPush  `cmd:"" help:"Push system image to registry"`
}

// CmdBuild 表示 build 命令参数。
type CmdBuild struct {
}

// Run 执行 build 命令。
func (cmd *CmdBuild) Run() error {
	build()
	return nil
}

// CmdPush 表示 push 命令参数。
type CmdPush struct {
	Image           string   `help:"Image reference (e.g. registry:5000/repo:tag)" required:""`
	Username        string   `help:"Registry username" env:"REGISTRY_USERNAME"`
	Password        string   `help:"Registry password" env:"REGISTRY_PASSWORD"`
	Insecure        bool     `help:"Use HTTP instead of HTTPS" default:"false"`
	Exclude         []string `help:"Exclude path from rootfs"`
	DefaultExcludes bool     `help:"Use default rootfs exclude paths" default:"false"`
}

// Run 执行 push 命令。
func (cmd *CmdPush) Run() error {
	return push(cmd)
}
