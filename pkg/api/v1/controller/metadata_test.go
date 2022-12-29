package controller

import (
	"testing"
)

func TestMetadataList(t *testing.T) {
	t.Run("TestMetadataList", func(t *testing.T) {
		//testMetadataStrList := []string{
		//	"bafy2bzacean4vsnuenpphxc3vgfbyhnsgfqfseafwnlpdjxebmgs46u23izv6",
		//	"bafy2bzacec5ogl7fg66g4qr4pd344exntv3j4iokr3mugx5a3v7syd3r2txcc",
		//	"bafy2bzacedavfzjsrskccma457nhnszykjtci6ae7lik276i5ziymxftiencw",
		//}
		//testMetadataList, err := cids.DecodeAndPadSnapShotList(testMetadataStrList)
		//if err != nil {
		//	t.Error(err)
		//}

		t.Run("when GetSnapShotCidList returns valid cid list, should returns that list", func(t *testing.T) {
			//ToDo mock dependencies
			//patch := gomonkey.ApplyMethodFunc(
			//	reflect.TypeOf(&snapshotstore.SnapShotStore{}),
			//	"GetSnapShotList",
			//	func(ctx context.Context) (*store.SnapShotList, error) {
			//		return testMetadataList, nil
			//	},
			//)
			//defer patch.Reset()
			//
			//actualMetaDataListBytes, err := mockController.SnapShotList()
			//if err != nil {
			//	t.Error(err)
			//}
			//var actualMetaDataList store.SnapShotList
			//err = json.Unmarshal(actualMetaDataListBytes, &actualMetaDataList)
			//if err != nil {
			//	t.Error(err)
			//}
			//So(&actualMetaDataList, ShouldResemble, testMetadataList)
		})

		t.Run("when GetSnapShotCidList returns an error, should returns that error with code 500", func(t *testing.T) {
			//patch := gomonkey.ApplyMethodFunc(
			//	reflect.TypeOf(&snapshotstore.SnapShotStore{}),
			//	"GetSnapShotList",
			//	func(context.Context) (*store.SnapShotList, error) {
			//		return nil, fmt.Errorf("monkey error")
			//	},
			//)
			//defer patch.Reset()
			//
			//var apiError *v1.Error
			//actualMetaDataListBytes, err := mockController.SnapShotList()
			//So(actualMetaDataListBytes, ShouldBeNil)
			//errors.As(err, &apiError)
			//So(apiError.Status(), ShouldEqual, http.StatusInternalServerError)
			//So(apiError.Error(), ShouldEqual, "monkey error")
		})

		t.Run("when GetSnapShotCidList returns a nil value and an nil error, should returns resource not found error with code 404", func(t *testing.T) {
			//patch := gomonkey.ApplyMethodFunc(
			//	reflect.TypeOf(&snapshotstore.SnapShotStore{}),
			//	"GetSnapShotList",
			//	func(context.Context) (*store.SnapShotList, error) {
			//		return nil, nil
			//	},
			//)
			//defer patch.Reset()
			//
			//var apiError *v1.Error
			//actualMetaDataListBytes, err := mockController.SnapShotList()
			//So(actualMetaDataListBytes, ShouldBeNil)
			//errors.As(err, &apiError)
			//So(apiError.Status(), ShouldEqual, http.StatusNotFound)
			//So(apiError.Error(), ShouldEqual, v1.ResourceNotFound.Error())
		})
	})
}
