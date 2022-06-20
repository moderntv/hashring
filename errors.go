package hashring

import "errors"

var (
	ErrNoNodes       = errors.New("no nodes available")
	ErrNodeNotFound  = errors.New("node not found")
	ErrDuplicateNode = errors.New("node already exists")
)
