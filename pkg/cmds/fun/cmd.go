package fun

import (
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/sunrise"
	"github.com/rwxrob/bonzai/comp"
)

var Cmd = &bonzai.Cmd{
	Name: `fun`,
	Cmds: []*bonzai.Cmd{sunrise.Cmd},
	Comp: comp.Cmds,
}
