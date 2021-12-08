package metadata

import (
	"Pando/internal/httpclient"
	"bytes"
	"encoding/json"
	"fmt"
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

const (
	checkInterval  = time.Second * 10
	fileDir        = "./tmp"
	shutGateway    = "https://shuttle-4.estuary.tech"
	estuaryGateway = "https://api.estuary.tech"

	EstuaryApiKeyEnv = "ESTKEY"
)

type addResponse struct {
	Cid       string
	EstuaryId uint64
	Providers []string
}

type contentStatusResponse struct {
	Content struct {
		Id           int    `json:"id"`
		Cid          string `json:"cid"`
		Name         string `json:"name"`
		UserId       int    `json:"userId"`
		Description  string `json:"description"`
		Size         int    `json:"size"`
		Active       bool   `json:"active"`
		Offloaded    bool   `json:"offloaded"`
		Replication  int    `json:"replication"`
		AggregatedIn int    `json:"aggregatedIn"`
		Aggregate    bool   `json:"aggregate"`
		Pinning      bool   `json:"pinning"`
		PinMeta      string `json:"pinMeta"`
		Failed       bool   `json:"failed"`
		Location     string `json:"location"`
		DagSplit     bool   `json:"dagSplit"`
	} `json:"content"`
	Deals         []interface{} `json:"deals"`
	FailuresCount int           `json:"failuresCount"`
}

type backupSystem struct {
	gateway        string
	shuttleGateway string
	checkInterval  time.Duration
	apiKey         string
	fileDir        string
	fileName       string
	toCheck        chan uint64
}

func NewBackupSys(gateway string, shuttleGateway string) (*backupSystem, error) {
	if gateway == "" {
		gateway = estuaryGateway
	}
	if shuttleGateway == "" {
		shuttleGateway = shutGateway
	}

	apiKey, exist := os.LookupEnv(EstuaryApiKeyEnv)
	if !exist {
		return nil, fmt.Errorf("please set apikey in $%s", EstuaryApiKeyEnv)
	}
	bs := &backupSystem{
		gateway:        gateway,
		shuttleGateway: shuttleGateway,
		checkInterval:  time.Second * 10,
		apiKey:         apiKey,
		fileDir:        fileDir,
		toCheck:        make(chan uint64, 1),
	}
	bs.run()

	return bs, nil
}

func (bs *backupSystem) run() {
	// if there is car file, back up it then deletes file
	go func() {
		for range time.NewTicker(time.Second).C {
			files, err := ioutil.ReadDir(BackupTmpPath)
			if err != nil {
				log.Error("wrong back up dir path: %s", BackupTmpPath)
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
	mux := new(sync.Mutex)
	go func() {
		for estId := range bs.toCheck {
			mux.Lock()
			waitCheckList = append(waitCheckList, estId)
			mux.Unlock()
		}
	}()

	for range time.NewTicker(checkInterval).C {
		for idx, checkId := range waitCheckList {
			success, err := bs.checkDealForBackup(checkId)
			if err != nil {
				log.Errorf("failed to check the status of content id : %d, err : %s", checkId, err.Error())
			}
			if success {
				mux.Lock()
				if waitCheckList[idx] == checkId {
					waitCheckList = append(waitCheckList[:idx], waitCheckList[idx+1:]...)
				} else {
					// maybe the waitCheckList is adding while checking status
					for i := idx + 1; i < len(waitCheckList); i++ {
						if waitCheckList[i] == checkId {
							waitCheckList = append(waitCheckList[:i], waitCheckList[i+1:]...)
							break
						}
					}
				}
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
	defer res.Body.Close()
	body, e := io.ReadAll(res.Body)
	if e != nil {
		log.Error("failed to read response body")
		return false, err
	}
	// wrong response
	if res.StatusCode != 200 {
		err = httpclient.ReadError(res.StatusCode, body)
		log.Error(err.Error())
		return false, err
	}

	resStruct := new(contentStatusResponse)
	err = json.Unmarshal(body, resStruct)
	if err != nil {
		return false, err
	}

	if resStruct.Deals == nil {
		return false, nil
	}

	return false, nil
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
		return err
	}
	io.Copy(formfile, bytes.NewReader(data))

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
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	r := new(addResponse)
	err = json.Unmarshal(body, r)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return httpclient.ReadError(res.StatusCode, body)
	}

	bs.toCheck <- r.EstuaryId
	log.Debugf("success back up %s to est, id : %d", filepath, r.EstuaryId)

	return nil
}
