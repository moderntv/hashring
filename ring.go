package hashring

import (
	"errors"
	"fmt"
	"hash"
	"sort"
	"sync"
)

var ErrNoNodes = errors.New("no nodes available")

type nodesSet []uint64

func (this nodesSet) Len() int           { return len(this) }
func (this nodesSet) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }
func (this nodesSet) Less(i, j int) bool { return this[i] < this[j] }

type Ring struct {
	sync.RWMutex

	virtualNodes        nodesSet
	virtualNodesMapping map[uint64]string

	hash                hash.Hash64
	virtualNodesPerNode uint
}

func NewRing(opts ...Option) (*Ring, error) {
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

func (this *Ring) getHash(key []byte) (uint64, error) {
	this.hash.Reset()
	_, err := this.hash.Write(key)
	if err != nil {
		return 0, err
	}

	return this.hash.Sum64(), nil
}

func (this *Ring) AddNode(node string) error {
	this.Lock()
	defer this.Unlock()

	for i := 0; uint(i) < this.virtualNodesPerNode; i++ {
		virtualNodeKey := fmt.Sprintf("%s:%d", node, i)
		virtualNodeHash, err := this.getHash([]byte(virtualNodeKey))
		if err != nil {
			return fmt.Errorf("cannot add node: %v", err)
		}

		this.virtualNodes = append(this.virtualNodes, virtualNodeHash)
		this.virtualNodesMapping[virtualNodeHash] = node
	}

	sort.Sort(this.virtualNodes)

	return nil
}

func (this *Ring) DeleteNode(node string) error {
	this.Lock()
	defer this.Unlock()

	for i := 0; uint(i) < this.virtualNodesPerNode; i++ {
		virtualNodeKey := fmt.Sprintf("%s:%d", node, i)
		virtualNodeHash, err := this.getHash([]byte(virtualNodeKey))
		if err != nil {
			return fmt.Errorf("cannot add node: %v", err)
		}

		delete(this.virtualNodesMapping, virtualNodeHash)
	}

	hashes := make(nodesSet, 0)
	for virtualNodeHash, _ := range this.virtualNodesMapping {
		hashes = append(hashes, virtualNodeHash)
	}
	this.virtualNodes = hashes
	sort.Sort(this.virtualNodes)

	return nil
}

func (this *Ring) GetNode(key string) (string, error) {
	this.RLock()
	defer this.RUnlock()

	if len(this.virtualNodes) <= 0 {
		return "", ErrNoNodes
	}

	hash, err := this.getHash([]byte(key))
	if err != nil {
		return "", err
	}

	p := sort.Search(len(this.virtualNodes), func(i int) bool { return this.virtualNodes[i] >= hash })
	if p == len(this.virtualNodes) {
		return this.virtualNodesMapping[this.virtualNodes[0]], nil
	}

	return this.virtualNodesMapping[this.virtualNodes[p]], nil
}
