package modify

func commands() string {
	return `package bootstrap

import "github.com/goravel/framework/contracts/console"

func Commands() []console.Command {
	return []console.Command{}
}
`
}
