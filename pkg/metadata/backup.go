package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kenlabs/pando/pkg/metadata/est_utils"
	"github.com/kenlabs/pando/pkg/option"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

var (
	CheckInterval         = time.Second * 10
	DefaultShuttleGateway = "https://shuttle-5.estuary.tech"
	DefaultEstGateway     = "https://api.estuary.tech"
)

type BackupSystem struct {
	backupCfg *option.Backup
	apiKey    string
	toCheck   chan uint64
}

func NewBackupSys(backupCfg *option.Backup) (*BackupSystem, error) {
	bs := &BackupSystem{
		apiKey:    "Bearer " + backupCfg.APIKey,
		toCheck:   make(chan uint64, 1),
		backupCfg: backupCfg,
	}
	err := bs.run()
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (bs *BackupSystem) run() error {
	// if there is car file, back up it then delete file
	backupInterval, err := time.ParseDuration(bs.backupCfg.BackupEstInterval)
	if err != nil {
		return err
	}
	checkInterval, err := time.ParseDuration(bs.backupCfg.EstCheckInterval)
	if err != nil {
		return err
	}
	go func() {
		for range time.NewTicker(backupInterval).C {
			files, err := ioutil.ReadDir(BackupTmpPath)
			if err != nil {
				logger.Errorf("wrong back up dir path: %s", BackupTmpPath)
			}
			for _, file := range files {
				if file.IsDir() {
					// dir should not back up
					continue
				}
				_, err = bs.BackupToEstuary(path.Join(BackupTmpPath, file.Name()))
				if err != nil {
					//todo metrics
					logger.Warnf("failed back up, err : %s", err.Error())
					continue
				}
				err = os.Remove(path.Join(BackupTmpPath, file.Name()))
				if err != nil {
					logger.Error("failed to remove the backed up car file")
				}
			}
		}
	}()

	go bs.checkDeal(checkInterval)

	return nil
}

func (bs *BackupSystem) checkDeal(checkInterval time.Duration) {
	waitCheckList := make([]uint64, 0)
	mux := sync.Mutex{}
	go func() {
		for estId := range bs.toCheck {
			mux.Lock()
			waitCheckList = append(waitCheckList, estId)
			mux.Unlock()
		}
	}()

	for range time.NewTicker(checkInterval).C {
		for idx, checkId := range waitCheckList {
			if len(waitCheckList) < idx+1 {
				continue
			}
			success, err := bs.checkDealForBackup(checkId)
			if err != nil {
				logger.Errorf("failed to check deal status of content id : %d, err : %s", checkId, err.Error())
				continue
			}
			// delete from waitCheckList
			if success {
				logger.Debugf("est : %d is successful to back up in filecoin!", checkId)
				mux.Lock()
				if waitCheckList[idx] == checkId {
					if len(waitCheckList) > 1 {
						waitCheckList = append(waitCheckList[:idx], waitCheckList[idx+1:]...)
					} else {
						waitCheckList = make([]uint64, 0)
					}
				} else {
					// maybe the waitCheckList is adding while checking status
					for i := idx + 1; i < len(waitCheckList); i++ {
						if waitCheckList[i] == checkId {
							waitCheckList = append(waitCheckList[:i], waitCheckList[i+1:]...)
							break
						}
					}
				}
				mux.Unlock()
			} else {
				// todo: metrics for failure
			}
		}
	}
}

func (bs *BackupSystem) checkDealForBackup(estID uint64) (bool, error) {
	req, err := http.NewRequest("GET", bs.backupCfg.EstuaryGateway+"/content/status/"+strconv.FormatUint(estID, 10), nil)
	if err != nil {
		logger.Error("failed to create request: %s", err.Error())
	}
	req.Header.Set("Authorization", bs.apiKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("failed to send http request: %s", err.Error())
		return false, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	body, e := io.ReadAll(res.Body)
	if e != nil {
		logger.Error("failed to read response body")
		return false, err
	}
	// wrong response
	if res.StatusCode != 200 {
		//err = httpclient.ReadError(res.StatusCode, body)
		//logger.Error(err.Error())
		return false, fmt.Errorf("fail response: %v", string(body))
	}

	resStruct := new(est_utils.ContentStatus)
	err = json.Unmarshal(body, resStruct)
	if err != nil {
		return false, err
	}

	success, err := bs.checkDealStatus(resStruct)
	if err != nil {
		logger.Errorf("failed to check the status of deal for file: %s, id: %d, err: %s", resStruct.Content.Name, estID, err.Error())
		return false, err
	}
	if success {
		return true, nil
	} else {
		return false, nil
	}
}

func (bs *BackupSystem) BackupToEstuary(filepath string) (uint64, error) {
	fBuf := new(bytes.Buffer)
	mw := multipart.NewWriter(fBuf)
	fpath := fmt.Sprintf(`%s`, filepath)
	formfile, err := mw.CreateFormFile("data", fpath)
	if err != nil {
		return 0, err
	}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return 0, fmt.Errorf("failed to read data: %s", err.Error())
	}
	_, err = io.Copy(formfile, bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("failed to copy data: %s", err.Error())
	}

	if err := mw.Close(); err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", bs.backupCfg.ShuttleGateway+"/content/add", fBuf)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", bs.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", mw.FormDataContentType())

	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	r := new(est_utils.AddResponse)
	err = json.Unmarshal(body, r)
	if err != nil {
		return 0, err
	}

	if res.StatusCode != 200 {
		return 0, fmt.Errorf("fail response: %v", string(body))
	}

	bs.toCheck <- r.EstuaryId
	logger.Infof("back up %s to est successfully, estid : %d at time: %s", filepath, r.EstuaryId, time.Now().String())

	return r.EstuaryId, nil
}

func (bs *BackupSystem) checkDealStatus(cs *est_utils.ContentStatus) (bool, error) {
	if cs == nil {
		return false, fmt.Errorf("nil conten status")
	}
	if len(cs.Deals) == 0 {
		return false, nil
	}

	for _, ds := range cs.Deals {
		if ds.Deal.Failed {
			return false, nil
		}
	}
	return true, nil
}
