package hashring

type nodesSet []uint64

func (ns nodesSet) Len() int           { return len(ns) }
func (ns nodesSet) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns nodesSet) Less(i, j int) bool { return ns[i] < ns[j] }
