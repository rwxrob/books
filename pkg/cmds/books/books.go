package books

import (
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
	"github.com/rwxrob/books/pkg/cmds/fun"
)

var Cmd = &bonzai.Cmd{
	Name:  `rwxrob-books`,
	Alias: `books`,
	Cmds:  []*bonzai.Cmd{help.Cmd, fun.Cmd},
	Comp:  comp.Cmds,
}
