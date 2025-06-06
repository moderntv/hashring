package hashring

import (
	"hash"
	"hash/fnv"
)

const (
	defaultVirtualNodesPerNode = 1000
)

type Option func(*options) error

type options struct {
	virtualNodesPerNode uint
	hash                hash.Hash64
}

func defaultOptions() options {
	return options{
		virtualNodesPerNode: defaultVirtualNodesPerNode,
		hash:                fnv.New64a(),
	}
}

func WithVirtualNodeReplicas(count uint) Option {
	return func(options *options) error {
		options.virtualNodesPerNode = count
		return nil
	}
}

func WithHashFunc(hash hash.Hash64) Option {
	return func(options *options) error {
		options.hash = hash
		return nil
	}
}
