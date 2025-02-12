// in replicate/interfaces.go
package replicate

// ClusterNodeProvider defines methods needed from GrpcServer
type ClusterNodeProvider interface {
	GetFollowerNodes() map[int32]string
}

// PersistenceProvider defines an interface for persistence operations
type PersistenceProvider interface {
	LoadRequests() ([]map[string]interface{}, error)
}
