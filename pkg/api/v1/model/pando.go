package model

type PandoInfo struct {
	PeerID    string
	Addresses APIAddresses
}

type APIAddresses struct {
	HttpAPI      string
	GraphQLAPI   string
	GraphSyncAPI string
}
