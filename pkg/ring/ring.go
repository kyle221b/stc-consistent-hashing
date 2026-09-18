package ring

func hsum(data string) uint32 {
	var sum int32
	for _, rune := range data {
		sum += rune
	}
	return uint32(sum)
}

func NewRing(keys []string, n uint32) *Ring {
	keyMapping := make(map[string]ShardId, len(keys))
	for _, key := range keys {
		keyMapping[key] = getShard(key, n)
	}
	return &Ring{
		ShardMapping: keyMapping,
		n:            n,
	}
}

type Ring struct {
	ShardMapping map[string]ShardId
	n            uint32
}

type ShardId uint32

func (r *Ring) GetShard(key string) uint32 {
	return uint32(r.ShardMapping[key])
}

func (r *Ring) Reshard(n uint32) int {
	if n == r.n {
		return 0
	}
	numChanged := 0
	for key, oldShardId := range r.ShardMapping {
		newShardId := getShard(key, n)
		r.ShardMapping[key] = newShardId
		if oldShardId != newShardId {
			numChanged++
		}
	}
	return numChanged
}

func getShard(key string, n uint32) ShardId {
	hash := hsum(key)
	return ShardId(hash % n)
}
