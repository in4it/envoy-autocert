package envoy

import "github.com/envoyproxy/go-control-plane/pkg/cache"

type WorkQueueItem struct {
	id               string
	Action           string
	DependsOn        string
	DependsOnItemIDs []string
	TLSParams        TLSParams
	ClusterParams    ClusterParams
	ListenerParams   ListenerParams
	ChallengeParams  ChallengeParams
	CreateCertParams CreateCertParams

	state string
}

type WorkQueueCache struct {
	snapshotCache cache.SnapshotCache
	clusters      []cache.Resource
	listeners     []cache.Resource
	version       int64
}

type WorkQueueSubmissionState struct {
	id     string
	state  string
	itemID string
}
type TLSParams struct {
	Name       string
	CertBundle string
	PrivateKey string
}
type ClusterParams struct {
	Name           string
	TargetHostname string
	Port           int64
}
type ListenerParams struct {
	Name           string
	Protocol       string
	TargetHostname string
	Conditions     Conditions
}

type ChallengeParams struct {
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	URI      string `json:"uri"`
	Token    string `json:"token"`
	Body     string `json:"body"`
	AuthzURI string `json:"authzURI"`
}
type CreateCertParams struct {
	Name            string
	Domains         []string
	DomainsToVerify []string
}
type Conditions struct {
	Hostname string
	Prefix   string
}