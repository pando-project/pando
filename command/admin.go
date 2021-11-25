package command

import (
	"Pando/internal/httpclient"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/urfave/cli/v2"
	"io"
	"net/http"
	"net/url"
)

const (
	adminPort      = 9002
	ingestResource = "graph"
)

var subscribe = &cli.Command{
	Name:   "subscribe",
	Usage:  "Subscribe provider with pando",
	Flags:  ingestFlags,
	Action: subscribeCmd,
}

var unsubscribe = &cli.Command{
	Name:   "unsubscribe",
	Usage:  "Unsubscribe provider from pando",
	Flags:  ingestFlags,
	Action: unsubscribeCmd,
}

var IngestCmd = &cli.Command{
	Name:  "ingest",
	Usage: "Admin commands to manage ingestion config of indexer",
	Subcommands: []*cli.Command{
		subscribe,
		unsubscribe,
	},
}

type ingestClient struct {
	c       *http.Client
	baseurl *url.URL
}

func subscribeCmd(cctx *cli.Context) error {
	err := sendRequest(cctx, "sub")
	if err != nil {
		return err
	}
	log.Infof("Successfully subscribed to provider")
	return nil
}

func unsubscribeCmd(cctx *cli.Context) error {
	err := sendRequest(cctx, "unsub")
	if err != nil {
		return err
	}
	log.Infof("Successfully unsubscribed from provider")
	return nil
}

func sendRequest(cctx *cli.Context, action string) error {
	cl, err := newIngestClient(cctx.String("pando"))
	if err != nil {
		return err
	}
	prov := cctx.String("provider")
	p, err := peer.Decode(prov)
	if err != nil {
		return err
	}
	dest := fmt.Sprintf("%s/%s/%s", cl.baseurl, action, p.String())
	req, err := http.NewRequestWithContext(cctx.Context, "GET", dest, nil)
	if err != nil {
		return err
	}

	resp, err := cl.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return httpclient.ReadError(resp.StatusCode, body)
	}
	return nil
}

func newIngestClient(baseurl string) (*ingestClient, error) {
	url, c, err := httpclient.New(baseurl, ingestResource, adminPort)
	if err != nil {
		return nil, err
	}
	return &ingestClient{
		c,
		url,
	}, nil
}
