package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"pando/pkg/metadata/est_utils"
	"pando/pkg/option"
	"path"
	"strconv"
	"sync"
	"time"
)

var (
	checkInterval         = time.Second * 10
	DefaultShuttleGateway = "https://shuttle-5.estuary.tech"
	DefaultEstGateway     = "https://api.estuary.tech"
)

type backupSystem struct {
	gateway        string
	shuttleGateway string
	checkInterval  time.Duration
	apiKey         string
	//fileDir        string
	fileName string
	toCheck  chan uint64
}

func NewBackupSys(backupCfg *option.Backup) (*backupSystem, error) {
	bs := &backupSystem{
		gateway:        backupCfg.EstuaryGateway,
		shuttleGateway: backupCfg.ShuttleGateway,
		checkInterval:  time.Second * 10,
		apiKey:         "Bearer " + backupCfg.APIKey,
		toCheck:        make(chan uint64, 1),
	}
	bs.run()

	return bs, nil
}

func (bs *backupSystem) run() {
	// if there is car file, back up it then delete file
	go func() {
		for range time.NewTicker(time.Second).C {
			files, err := ioutil.ReadDir(BackupTmpPath)
			if err != nil {
				log.Errorf("wrong back up dir path: %s", BackupTmpPath)
			}
			for _, file := range files {
				err = bs.backupToEstuary(path.Join(BackupTmpPath, file.Name()))
				if err != nil {
					//todo metrics
					log.Errorf("failed back up, err : %s", err.Error())
					continue
				}
				err = os.Remove(path.Join(BackupTmpPath, file.Name()))
				if err != nil {
					log.Error("failed to remove the backed up car file")
				}
			}
		}
	}()

	go bs.checkDeal()

}

func (bs *backupSystem) checkDeal() {
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
				log.Errorf("failed to check deal status of content id : %d, err : %s", checkId, err.Error())
				continue
			}
			// delete from waitCheckList
			if success {
				log.Debugf("est : %d is successful to back up in filecoin!", checkId)
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

func (bs *backupSystem) checkDealForBackup(estID uint64) (bool, error) {
	req, err := http.NewRequest("GET", bs.gateway+"/content/status/"+strconv.FormatUint(estID, 10), nil)
	if err != nil {
		log.Error("failed to create request: %s", err.Error())
	}
	req.Header.Set("Authorization", bs.apiKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("failed to send http request: %s", err.Error())
		return false, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	body, e := io.ReadAll(res.Body)
	if e != nil {
		log.Error("failed to read response body")
		return false, err
	}
	// wrong response
	if res.StatusCode != 200 {
		//err = httpclient.ReadError(res.StatusCode, body)
		//log.Error(err.Error())
		return false, fmt.Errorf("fail response: %v", string(body))
	}

	resStruct := new(est_utils.ContentStatus)
	err = json.Unmarshal(body, resStruct)
	if err != nil {
		return false, err
	}

	success, err := bs.checkDealStatus(resStruct)
	if err != nil {
		log.Errorf("failed to check the status of deal for file: %s, id: %d, err: %s", resStruct.Content.Name, estID, err.Error())
		return false, err
	}
	if success {
		return true, nil
	} else {
		return false, nil
	}
}

func (bs *backupSystem) backupToEstuary(filepath string) error {
	fBuf := new(bytes.Buffer)
	mw := multipart.NewWriter(fBuf)
	fpath := fmt.Sprintf(`%s`, filepath)
	formfile, err := mw.CreateFormFile("data", fpath)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read data: %s", err.Error())
	}
	_, err = io.Copy(formfile, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to copy data: %s", err.Error())
	}

	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", bs.shuttleGateway+"/content/add", fBuf)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", bs.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", mw.FormDataContentType())

	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	r := new(est_utils.AddResponse)
	err = json.Unmarshal(body, r)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("fail response: %v", string(body))
	}

	bs.toCheck <- r.EstuaryId
	log.Debugf("success back up %s to est, id : %d", filepath, r.EstuaryId)

	return nil
}

func (bs *backupSystem) checkDealStatus(cs *est_utils.ContentStatus) (bool, error) {
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
