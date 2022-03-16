package admin

import (
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/kenlabs/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
)

const backupPath = "/backup"

type backupReq struct {
	StartCid string
	EndCid   string
	Provider string
	IsForce  bool
}

var backupRequest = &backupReq{}

func backupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "backup metadata for Provider to estuary",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := backupRequest.validateFlags(); err != nil {
				return err
			}

			req := api.Client.R()
			req = req.SetQueryParam("start", backupRequest.StartCid)
			req = req.SetQueryParam("end", backupRequest.EndCid)
			var isForce string
			if backupRequest.IsForce == true {
				isForce = "1"
			} else {
				isForce = "0"
			}
			req = req.SetQueryParam("force", isForce)
			if backupRequest.Provider != "" {
				req = req.SetQueryParam("provider", backupRequest.Provider)
			}

			res, err := req.Get(joinAPIPath(backupPath))
			if err != nil {
				return err
			}
			return api.PrintResponseData(res)
		},
	}

	backupRequest.setFlags(cmd)

	return cmd
}

func (bq *backupReq) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&bq.StartCid, "startcid", "s", "",
		"the start cid for backup(not included in the car file)")
	cmd.Flags().StringVarP(&bq.EndCid, "endcid", "e", "",
		"the end cid for backup")
	cmd.Flags().StringVarP(&bq.Provider, "backup-provider", "p", "",
		"which provider the backup is for")
	cmd.Flags().BoolVarP(&bq.IsForce, "isforce", "f", false,
		"whether back up meta into estuary right now")
}

func (bq *backupReq) validateFlags() error {
	if bq.EndCid == "" {
		return fmt.Errorf("end cid can not be empty")
	}

	if _, err := cid.Decode(bq.EndCid); err != nil {
		return fmt.Errorf("invalid cid: %v", err)
	}

	if bq.StartCid != "" {
		if _, err := cid.Decode(bq.StartCid); err != nil {
			return fmt.Errorf("invalid cid: %v", err)
		}
	} else {
		bq.StartCid = bq.EndCid
	}

	return nil
}
