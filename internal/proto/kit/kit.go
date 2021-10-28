package kit

import (
	coresproto "github.com/doublemo/baa/cores/proto"
)

const (
	Agent coresproto.Command = 1 + (iota * 1000)
	SFU
	Auth
	SNID
	IM
	IMF
	USRT
)
