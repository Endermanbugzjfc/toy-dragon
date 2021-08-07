package systems

import "sync"

var Sessions sync.Map

type PlayerSession struct {
	Name *string

	JoinTime int
	Note     string
}
