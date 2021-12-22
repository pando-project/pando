package adminhttpclient

import (
	"Pando/api/v0/admin/model"
	"Pando/internal/httpclient"
	"bytes"
	"context"
	"io"
	"net/http"

	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	adminPort     = 9001
	providersPath = "/providers/register"
)

// Client is an http client for the indexer ingest API
type Client struct {
	c *http.Client
	//indexContentURL string
	providersURL string
}

// New creates a new ingest http Client
func New(baseURL string, options ...httpclient.Option) (*Client, error) {
	u, c, err := httpclient.New(baseURL, "", adminPort, options...)
	if err != nil {
		return nil, err
	}
	baseURL = u.String()
	return &Client{
		c: c,
		//indexContentURL: baseURL + indexContentPath,
		providersURL: baseURL + providersPath,
	}, nil
}

func (c *Client) Register(ctx context.Context, providerID peer.ID, privateKey p2pcrypto.PrivKey, addrs []string, miner string) error {
	data, err := model.MakeRegisterRequest(providerID, privateKey, addrs, miner)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.providersURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return httpclient.ReadError(resp.StatusCode, body)
	}
	return nil
}
