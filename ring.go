package hashring

import (
	"fmt"
	"hash"
	"sort"
	"sync"
)

type Ring struct {
	mu sync.RWMutex

	hashMu              sync.Mutex
	hash                hash.Hash64
	virtualNodesPerNode uint

	// nodes contains all nodes currently in hashring with backreferences to their virtual nodes
	nodes map[string]*node
	// virtualNodes is a list of virtual nodes hashes
	// this is the actual hashring
	virtualNodes nodesSet
	// virtualNodesMapping is a virtualNode -> nodeKey mapping
	virtualNodesMapping map[uint64]string
}

func New(opts ...Option) (*Ring, error) {
	options := defaultOptions()
	for _, modifier := range opts {
		if err := modifier(&options); err != nil {
			return nil, err
		}
	}

	r := Ring{
		virtualNodesPerNode: options.virtualNodesPerNode,
		hash:                options.hash,

		nodes:               map[string]*node{},
		virtualNodes:        make(nodesSet, 0),
		virtualNodesMapping: map[uint64]string{},
	}

	return &r, nil
}

func (r *Ring) AddNode(nodeKey string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	node := &node{
		key:          nodeKey,
		virtualNodes: []uint64{},
	}

	for i := range r.virtualNodesPerNode {
		virtualNodeKey := fmt.Sprintf("%s:%d", nodeKey, i)
		virtualNodeHash, err := r.getHash([]byte(virtualNodeKey))
		if err != nil {
			return fmt.Errorf("cannot add node: %w", err)
		}

		r.virtualNodes = append(r.virtualNodes, virtualNodeHash)
		r.virtualNodesMapping[virtualNodeHash] = nodeKey

		node.virtualNodes = append(node.virtualNodes, virtualNodeHash)
	}

	r.nodes[nodeKey] = node

	sort.Sort(r.virtualNodes)

	return nil
}

func (r *Ring) DeleteNode(nodeKey string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, ok := r.nodes[nodeKey]
	if !ok {
		return ErrNodeNotFound
	}

	for _, virtualNodeKey := range node.virtualNodes {
		delete(r.virtualNodesMapping, virtualNodeKey)
	}
	delete(r.nodes, nodeKey)

	// recreate THE hashring
	hashes := make(nodesSet, 0)
	for virtualNodeHash := range r.virtualNodesMapping {
		hashes = append(hashes, virtualNodeHash)
	}
	r.virtualNodes = hashes
	sort.Sort(r.virtualNodes)

	return nil
}

func (r *Ring) GetNode(key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.virtualNodes) == 0 {
		return "", ErrNoNodes
	}

	hash, err := r.getHash([]byte(key))
	if err != nil {
		return "", err
	}

	p := sort.Search(len(r.virtualNodes), func(i int) bool { return r.virtualNodes[i] >= hash })
	if p == len(r.virtualNodes) {
		return r.virtualNodesMapping[r.virtualNodes[0]], nil
	}

	return r.virtualNodesMapping[r.virtualNodes[p]], nil
}

func (r *Ring) getHash(key []byte) (uint64, error) {
	r.hashMu.Lock()
	defer r.hashMu.Unlock()

	r.hash.Reset()
	_, err := r.hash.Write(key)
	if err != nil {
		return 0, err
	}

	return r.hash.Sum64(), nil
}
