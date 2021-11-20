package speedtester

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/montanaflynn/stats"
	"github.com/shopspring/decimal"
	"github.com/showwin/speedtest-go/speedtest"
	"sort"
	"strings"
	"time"
)

func checkError(err error) {
	if err != nil {
		panic(fmt.Errorf("speedtester failed, error: %v", err))
	}
}

func FetchInternetSpeed() float64 {
	user, err := speedtest.FetchUserInfo()
	checkError(err)
	serverList, err := speedtest.FetchServerList(user)
	checkError(err)
	targets := serverList.Servers

	var downloadSpeedList []float64
	targetCount := len(targets)
	spin := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	spin.Prefix = "\ttesting"
	fmt.Printf("start speed test on %d targets, this step may take a while...\n", targetCount)
	for i, target := range targets {
		fmt.Printf("[%d/%d]test target server - %s in %s:\n", i+1, len(targets), target.Host, target.Country)
		// skip abroad target
		if serverIsAbroad(user.Country, target.Country) {
			fmt.Printf("\tskip abroad target\n")
			targetCount -= 1
			continue
		}
		spin.Start()
		err = target.DownloadTest(false)
		if err != nil {
			spin.Stop()
			fmt.Printf("\tdownload test failed, skip this target\n")
			targetCount -= 1
			continue
		}
		downloadSpeedList = append(downloadSpeedList, target.DLSpeed)
		spin.Stop()

		fmt.Printf("\tdownload speed = %.2f Mbps\n", target.DLSpeed)
	}

	sort.Float64s(downloadSpeedList)
	medianSpeed, err := stats.Median(downloadSpeedList)
	medianSpeed, _ = decimal.NewFromFloat(medianSpeed).Round(2).Float64()
	fmt.Printf("speed test complete, median download speed is %f Mbps\n", medianSpeed)
	return medianSpeed
}

func serverIsAbroad(userCountry string, serverCountry string) bool {
	userCountrySplit := strings.Split(strings.ToLower(userCountry), "")
	for _, countryCh := range userCountrySplit {
		if !strings.ContainsAny(strings.ToLower(serverCountry), countryCh) {
			return true
		}
	}
	return false
}
