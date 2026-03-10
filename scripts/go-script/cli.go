package main

var cli struct {
	E2E cmdE2e `cmd:"" name:"e2e" help:"Run e2e test"`
}

type cmdE2e struct {
}

func (c *cmdE2e) Run() error {
	e2e()
	return nil
}
