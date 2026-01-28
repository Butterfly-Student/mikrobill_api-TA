package redis_outbound_adapter

// Cache Key Templates
const (
	// PPP Resources (dengan tenant isolation)
	CachePPPProfilesKey = "ppp:profiles:%s" // %s = tenantId
	CachePPPSecretsKey  = "ppp:secrets:%s"  // %s = tenantId
	CachePPPActiveKey   = "ppp:active:%s"   // %s = tenantId
	CachePPPInactiveKey = "ppp:inactive:%s" // %s = tenantId

	// Pools and Queues
	CacheIPPoolsKey = "pools:%s"  // %s = tenantId
	CacheQueuesKey  = "queues:%s" // %s = tenantId

	// Hash keys untuk change detection (stored alongside data)
	CacheHashSuffix = ":hash"

	// PubSub Channels untuk real-time updates
	PubSubPPPActiveChannel   = "ppp:active:%s"   // %s = tenantId
	PubSubPPPInactiveChannel = "ppp:inactive:%s" // %s = tenantId
)
