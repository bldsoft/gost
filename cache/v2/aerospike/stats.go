package aerospike

type Stats struct {
	InPool            int `json:"inPool" mapstructure:"inPool"`
	OutPool           int `json:"outPool" mapstructure:"outPool"`
	OpenConns         int `json:"openConns" mapstructure:"openConns"`
	ConnectionsOpened int `json:"connectionsOpened" mapstructure:"connectionsOpened"`
	ConnectionsClosed int `json:"connectionsClosed" mapstructure:"connectionsClosed"`
	Nodes             int `json:"nodes" mapstructure:"nodes"`
	PartitionMapSize  int `json:"partitionMapSize" mapstructure:"partitionMapSize"`
	TendInterval      int `json:"tendInterval" mapstructure:"tendInterval"`
	NodesAdded        int `json:"nodesAdded" mapstructure:"nodesAdded"`
	NodesRemoved      int `json:"nodesRemoved" mapstructure:"nodesRemoved"`
	ErrorCount        int `json:"errorCount" mapstructure:"errorCount"`
}
