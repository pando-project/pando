package main

import (
	"context"
	"fmt"
	dt "github.com/filecoin-project/go-data-transfer"
	"github.com/goombaio/namegenerator"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/fluent/qp"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/pando-project/pando/pkg/types/schema"
	peerHelper "github.com/pando-project/pando/pkg/util/peer"
	pandoSdk "github.com/pando-project/pando/sdk/pkg/provider"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	privateKeyStr = "CAESQHWlReUYxW7FDvTAAqG+kNH2U7khW+iv0r+070+zKmFn9t80v5e30/NsBx5XzBLCE4uH/h3d3tpXlwCuO4YGN+w="
	pandoAddr     = "/ip4/127.0.0.1/tcp/9002"
	pandoPeerID   = "12D3KooWD67T3kwPSeJx7dc6wpxS6PZYHJkkjM4CeCMz1FSCT4y1"
	round         = 10000
)

var logger = logging.Logger("throughput_test")

func main() {
	logging.SetAllLoggers(logging.LevelInfo)
	err := logging.SetLogLevel("registry", "warn")
	if err != nil {
		panic(err)
	}
	peerID, err := peerHelper.GetPeerIDFromPrivateKeyStr(privateKeyStr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("provider peerID: %s", peerID.String())

	provider, err := pandoSdk.NewMetaProvider(
		privateKeyStr, "http://127.0.0.1:9000",
		10*time.Second, 10*time.Minute,
	)
	if err != nil {
		panic(err)
	}

	provider.DtManager.SubscribeToEvents(onDataTransferComplete)

	err = provider.ConnectPando(pandoAddr, pandoPeerID)
	time.Sleep(

		5 * time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Pando node: %s connected\n", pandoAddr)

	//for i := 0; i < round; i++ {
	//	metadata, err := generateRandomData(provider, strconv.Itoa(i))
	//	if err != nil {
	//		panic(err)
	//	}
	//	metadataCid, err := provider.Push(metadata)
	//	if err != nil {
	//		i--
	//		continue
	//	}
	//	fmt.Printf("%d metadata pushed, cid => %s\n", i+1, metadataCid)
	//	time.Sleep(5 * time.Second)
	//}
	root, err := generateRandomDataWithLink(provider, round)
	if err != nil {
		panic(err)
	}
	rootCid, err := provider.Push(root)
	fmt.Printf("%d metadata pushed, root cid: %s", round, rootCid.String())

	fmt.Println("press ctrl+c to exit.")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down provider...")
	err = provider.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println("Bye! ")
}

func generateRandomData(provider *pandoSdk.MetaProvider, id string) (*schema.Metadata, error) {
	seed := time.Now().UTC().UnixNano()
	ng := namegenerator.NewNameGenerator(seed)
	payloadNode, err := qp.BuildMap(basicnode.Prototype.Any, 4, func(ma datamodel.MapAssembler) {
		qp.MapEntry(ma, "id", qp.String(id))
		qp.MapEntry(ma, "name", qp.String(ng.Generate()))
		rand.Seed(seed)
		qp.MapEntry(ma, "age", qp.Int(int64(rand.Intn(120)+1)))
	})
	if err != nil {
		return nil, err
	}

	cache := true
	collection := "testNameList"
	return provider.NewMetadataWithPayload(payloadNode, &cache, &collection)
}

func generateRandomDataWithLink(provider *pandoSdk.MetaProvider, nodeQty int) (*schema.Metadata, error) {
	cache := true
	collection := "testNameListWithLink"
	seed := time.Now().UTC().UnixNano()
	ng := namegenerator.NewNameGenerator(seed)

	var metadata *schema.Metadata
	var prevCid cid.Cid
	var prevMetadata cidlink.Link
	for i := 0; i < nodeQty; i++ {
		payloadNode, err := qp.BuildMap(basicnode.Prototype.Any, 4, func(ma datamodel.MapAssembler) {
			qp.MapEntry(ma, "id", qp.String(strconv.Itoa(i)))
			qp.MapEntry(ma, "name", qp.String(ng.Generate()))
			rand.Seed(seed)
			qp.MapEntry(ma, "age", qp.Int(int64(rand.Intn(120)+1)))
		})
		if err != nil {
			return nil, err
		}

		if i == 0 {
			metadata, err = provider.NewMetadataWithPayload(payloadNode, &cache, &collection)
		} else {
			metadata, err = provider.NewMetadataWithPayloadLink(payloadNode, &cache, &collection, prevMetadata)
		}
		if err != nil {
			panic(err)
		}

		if i < nodeQty-1 {
			prevCid, err = provider.PushLocal(context.Background(), metadata)
			if err != nil {
				return metadata, err
			}
			prevMetadata = cidlink.Link{Cid: prevCid}
			logger.Infof("[%.2f%%] %d metadata pushed in local store, cid: %s\n", (float64(i+1)/float64(nodeQty))*100, i+1, prevCid)
		}
	}
	return metadata, nil
}

var (
	startTime  time.Time
	timeTicker *time.Ticker
	count      = 0
	countLock  = sync.RWMutex{}
	lastCount  = 0
)

func onDataTransferComplete(event dt.Event, channelState dt.ChannelState) {
	logger.Debugf("transfer event: %s, cid: %s\n", dt.Events[event.Code], channelState.BaseCID())

	if event.Code == dt.Open {
		startTime = time.Now()
		timeTicker = time.NewTicker(time.Second)
		go iopsMonitor(round)
	}
	if event.Code == dt.DataSent {
		countLock.Lock()
		count++
		countLock.Unlock()
	}
	if event.Code == dt.Complete {
		timeTicker.Stop()
		intervalSeconds := math.Ceil(-1 * time.Until(startTime).Seconds())
		aveIOPS := count / int(intervalSeconds)
		logger.Warnf("[Average IOPS] %d ops/s, transffered %d records", aveIOPS, count)
	}
}

func iopsMonitor(round int) {
	for {
		select {
		case <-timeTicker.C:
			countLock.Lock()
			if count < round {
				roundCount := count - lastCount
				lastCount = count
				logger.Warnf("[IOPS] %d ops/s, transferred %d records", roundCount, count)
			} else {
				timeTicker.Stop()
			}
			countLock.Unlock()
		}
	}
}

func appendLog(entry string) {
	fName := "./throughput.log"
	f, err := os.OpenFile(fName, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString(entry)
	if err != nil {
		panic(err)
	}
}
