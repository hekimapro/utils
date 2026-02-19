package snowflake

import (
	"sync"
	"time"

	"github.com/hekimapro/utils/helpers"
)

const (
	nodeBits     = 10
	sequenceBits = 12

	maxNodeID   = -1 ^ (-1 << nodeBits)
	maxSequence = -1 ^ (-1 << sequenceBits)

	nodeShift = sequenceBits
	timeShift = sequenceBits + nodeBits

	// Twitter original epoch: Nov 4, 2010
	epoch int64 = 1288834974657
)

type generator struct {
	mu        sync.Mutex
	nodeID    int64
	sequence  int64
	lastStamp int64
}

var g = newGenerator()

func newGenerator() *generator {
	nodeID := helpers.GetENVIntValue("SNOWFLAKE_NODE_ID", 0)

	if nodeID < 0 {
		nodeID = 0
	}

	if nodeID > maxNodeID {
		nodeID = int(maxNodeID)
	}

	return &generator{
		nodeID: int64(nodeID),
	}
}

func NextID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	ts := time.Now().UnixMilli()

	if ts == g.lastStamp {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			for ts <= g.lastStamp {
				ts = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastStamp = ts

	return ((ts - epoch) << timeShift) |
		(g.nodeID << nodeShift) |
		g.sequence
}
