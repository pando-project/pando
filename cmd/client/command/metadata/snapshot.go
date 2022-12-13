package metadata

import (
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/pando-project/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
	"strconv"
)

const snapshotPath = "/snapshot"

type snapshotAPIQuery struct {
	cid    string
	height string
}

var snapshotQuery = &snapshotAPIQuery{}

func snapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "lookup a snapshot with specified snapshot cid or its height",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := snapshotQuery.validateFlags(); err != nil {
				return err
			}

			req := api.Client.R()
			if snapshotQuery.cid != "" {
				req = req.SetQueryParam("cid", snapshotQuery.cid)
			} else if snapshotQuery.height != "" {
				req = req.SetQueryParam("height", snapshotQuery.height)
			}

			res, err := req.Get(joinAPIPath(snapshotPath))
			if err != nil {
				return err
			}
			return api.PrintResponseData(res)
		},
	}
	snapshotQuery.setFlags(cmd)

	return cmd
}

func (q *snapshotAPIQuery) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&q.cid, "snapshot-cid", "c", "",
		"lookup a snapshot by its cid")
	cmd.Flags().StringVarP(&q.height, "snapshot-height", "t", "",
		"lookup a snapshot by its height")
}

func (q *snapshotAPIQuery) validateFlags() error {
	if q.cid != "" {
		_, err := cid.Decode(q.cid)
		if err != nil {
			return fmt.Errorf("invalid cid: %v", err)
		} else {
			return nil
		}
	} else if q.height != "" {
		_, err := strconv.ParseUint(q.height, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid height: %v", err)
		}
	} else {
		return fmt.Errorf("either cid or height should be specified to lookup a snapshot")
	}

	return nil
}
