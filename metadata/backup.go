package metadata

import (
	"Pando/internal/httpclient"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	urlpkg "net/url"
	"os"
	"path"
	"time"
)

const (
	checkInterval = time.Hour * 1
	fileDir       = "./tmp"
	filename      = "backup.car"
)

type backupSystem struct {
	gateway       string
	checkInterval time.Duration
	apiKey        string
	fileDir       string
	fileName      string
	//c             *http.Client
}

func NewBackupSys(gateway string, apiKey string) (*backupSystem, error) {
	//c, err := http.Client{}

	return &backupSystem{
		gateway:       gateway,
		checkInterval: time.Second * 10,
		apiKey:        apiKey,
		fileDir:       fileDir,
		fileName:      filename,
	}, nil
}

func (bs *backupSystem) run() {
	for range time.NewTicker(time.Second).C {
		if _, err := os.Stat(path.Join(fileDir, filename)); err == nil {
			// todo back
		}
	}
}

//
//
//# curl -X POST http://localhost:3004/content/add
//#   -H "Authorization: Bearer ESTfdd10399-b542-4274-bea5-dc12068ff2d6ARY"
//#   -H "Accept: application/json"
//#   -H "Content-Type: multipart/form-data"
//#   -F "data=@/Users/zxh/Code/estuary/ranking.go"
//POST https://shuttle-4.estuary.tech/content/add
//Authorization: Bearer EST75c4d3bb-d86f-42e4-80da-662d7fbde4c2ARY
//Accept: application/json
//Content-Type: multipart/form-data; boundary=WebAppBoundary
//
//--WebAppBoundary
//Content-Disposition: form-data; name="data"; filename="ranking.go"
//
//< /Users/zxh/Code/estuary/ranking.go
//--WebAppBoundary--
//

func (bs *backupSystem) backupToEstuary(filepath string) error {
	req := http.Request{
		Method: "POST",
		//Header: map[string][]string{},
		Form: map[string][]string{},
	}
	req.Header = map[string][]string{
		//"Content-Type": {"multipart/form-data"},
	}

	//_, err := req.MultipartReader()
	//if err != nil {
	//	return err
	//}
	req.Header.Set("Authorization", bs.apiKey)
	req.Header.Set("Accept", "application/json")

	req.Form.Add("data", filepath)

	url, err := urlpkg.Parse(bs.gateway + "/content/add")
	if err != nil {
		return err
	}
	req.URL = url
	res, err := (&http.Client{}).Do(&req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return httpclient.ReadError(res.StatusCode, body)
	}
	return nil
}

func (bs *backupSystem) backupToEstuary_(filepath string) error {
	fBuf := new(bytes.Buffer)
	mw := multipart.NewWriter(fBuf)
	fpath := fmt.Sprintf(`%s`, filepath)
	err := mw.WriteField("data", fpath)
	if err != nil {
		return err
	}

	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", bs.gateway+"/content/add", fBuf)
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

	if res.StatusCode != 200 {
		return httpclient.ReadError(res.StatusCode, body)
	}
	return nil
}
