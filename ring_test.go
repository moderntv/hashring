package hashring //nolint: testpackage

import (
	"fmt"
	"log"
	"math"
	"testing"

	"github.com/cespare/xxhash/v2"
	"github.com/matryer/is"
)

const numTestKeys = 1_000_000

type perturbType int

const (
	add perturbType = iota
	remove
)

func TestHashring(t *testing.T) {
	t.Parallel()

	is := is.New(t)

	tests := []struct {
		name              string
		replicationFactor uint
		nodes             []string
	}{
		{
			name:              "simple",
			replicationFactor: 3,
			nodes:             []string{"node1"},
		},
		{
			name:              "multiple nodes",
			replicationFactor: 10000,
			nodes:             []string{"node1", "node2", "node3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := is.New(t)

			ring, err := New(WithVirtualNodeReplicas(tt.replicationFactor), WithHashFunc(xxhash.New()))
			is.NoErr(err)

			for _, node := range tt.nodes {
				err := ring.AddNode(node)
				is.NoErr(err)
			}

			is.Equal(len(ring.nodes), len(tt.nodes))
			is.Equal(uint(len(ring.virtualNodes)), uint(len(tt.nodes))*tt.replicationFactor)

			for _, ringNode := range ring.nodes {
				is.Equal(uint(len(ringNode.virtualNodes)), tt.replicationFactor)
			}

			for _, node := range tt.nodes {
				err := ring.DeleteNode(node)
				is.NoErr(err)
			}

			is.Equal(len(ring.nodes), 0)
			is.Equal(len(ring.virtualNodes), 0)
		})
	}
}

func TestBalance(t *testing.T) {
	t.Parallel()

	is := is.New(t)

	ring, err := New(WithHashFunc(xxhash.New()))
	is.NoErr(err)

	nodes := []string{"node1", "node2", "node3"}
	mappings := map[string]int{}
	for _, node := range nodes {
		err := ring.AddNode(node)
		is.NoErr(err)
		mappings[node] = 0
	}

	for k := range numTestKeys {
		node, err := ring.GetNode(fmt.Sprintf("key%d", k))
		is.NoErr(err)

		mappings[node]++
	}

	total := 0
	stddevSum := 0.0
	mean := float64(numTestKeys) / float64(len(nodes))
	for node, nodeTotal := range mappings {
		log.Printf("%s: %d", node, nodeTotal)
		total += nodeTotal
		stddevSum += math.Pow(float64(nodeTotal)-mean, 2)
	}

	stddev := math.Sqrt(stddevSum / float64(len(nodes)))

	is.True(stddev < mean*0.1)
}

func TestConsistency(t *testing.T) {
	t.Parallel()

	is := is.New(t)

	ring, err := New(WithHashFunc(xxhash.New()))
	is.NoErr(err)

	nodes := []string{"node1", "node2", "node3", "node4", "node5"}
	for _, node := range nodes {
		err := ring.AddNode(node)
		is.NoErr(err)
	}

	mappings := map[string]string{}
	for k := range numTestKeys {
		key := fmt.Sprintf("key%d", k)
		node, err := ring.GetNode(key)
		is.NoErr(err)

		mappings[key] = node
	}

	err = ring.DeleteNode("node2")
	is.NoErr(err)
	mappings = checkValid(is, ring, "node2", add, mappings)

	err = ring.DeleteNode("node5")
	is.NoErr(err)
	mappings = checkValid(is, ring, "node5", add, mappings)

	err = ring.AddNode("node2")
	is.NoErr(err)

	mappings = checkValid(is, ring, "node2", add, mappings)
	_ = mappings
}

func checkValid(
	is *is.I,
	ring *Ring,
	changedNode string,
	typ perturbType,
	previousMapping map[string]string,
) (newMapping map[string]string) {
	is.Helper()

	newMapping = map[string]string{}

	for k := range numTestKeys {
		key := fmt.Sprintf("key%d", k)

		node, err := ring.GetNode(key)
		is.NoErr(err)

		newMapping[key] = node

		if (typ == remove && previousMapping[key] == changedNode) || (typ == add && newMapping[key] == changedNode) {
			continue
		}

		is.Equal(node, newMapping[key])
	}

	return
}
