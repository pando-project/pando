package sdk

type Provider interface {
	ConnectPando()
	Close()
	Push()
}
