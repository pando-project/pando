package provider

import ipldFormat "github.com/ipfs/go-ipld-format"

type BlockProvider struct {
}

func NewBlockProvider() *BlockProvider {

	return &BlockProvider{}
}

func (p *BlockProvider) ConnectPando(peerAddress string, peerID string) error {

	return nil
}

func (p *BlockProvider) Close() error {

	return nil
}

func (p *BlockProvider) Push(node ipldFormat.Node) error {

	return nil
}
