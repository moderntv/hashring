package hashring

import (
	"fmt"
	"hash"
	"sort"
	"sync"
)

type Ring struct {
	sync.RWMutex

	virtualNodes        nodesSet
	virtualNodesMapping map[uint64]string

	hash                hash.Hash64
	virtualNodesPerNode uint
}

func New(opts ...Option) (*Ring, error) {
	options := defaultOptions()
	for _, modifier := range opts {
		if err := modifier(&options); err != nil {
			return nil, err
		}
	}

	r := Ring{
		virtualNodes:        make(nodesSet, 0),
		virtualNodesMapping: map[uint64]string{},

		virtualNodesPerNode: options.virtualNodesPerNode,
		hash:                options.hash,
	}

	return &r, nil
}

func (r *Ring) AddNode(node string) error {
	r.Lock()
	defer r.Unlock()

	for i := 0; uint(i) < r.virtualNodesPerNode; i++ {
		virtualNodeKey := fmt.Sprintf("%s:%d", node, i)
		virtualNodeHash, err := r.getHash([]byte(virtualNodeKey))
		if err != nil {
			return fmt.Errorf("cannot add node: %v", err)
		}

		r.virtualNodes = append(r.virtualNodes, virtualNodeHash)
		r.virtualNodesMapping[virtualNodeHash] = node
	}

	sort.Sort(r.virtualNodes)

	return nil
}

func (r *Ring) DeleteNode(node string) error {
	r.Lock()
	defer r.Unlock()

	for i := 0; uint(i) < r.virtualNodesPerNode; i++ {
		virtualNodeKey := fmt.Sprintf("%s:%d", node, i)
		virtualNodeHash, err := r.getHash([]byte(virtualNodeKey))
		if err != nil {
			return fmt.Errorf("cannot add node: %v", err)
		}

		delete(r.virtualNodesMapping, virtualNodeHash)
	}

	hashes := make(nodesSet, 0)
	for virtualNodeHash := range r.virtualNodesMapping {
		hashes = append(hashes, virtualNodeHash)
	}
	r.virtualNodes = hashes
	sort.Sort(r.virtualNodes)

	return nil
}

func (r *Ring) GetNode(key string) (string, error) {
	r.RLock()
	defer r.RUnlock()

	if len(r.virtualNodes) <= 0 {
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
	r.hash.Reset()
	_, err := r.hash.Write(key)
	if err != nil {
		return 0, err
	}

	return r.hash.Sum64(), nil
}
