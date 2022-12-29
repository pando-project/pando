package pando

import (
	"testing"
)

func TestProviderRegister(t *testing.T) {
	t.Run("TestProviderRegister", func(t *testing.T) {
		//responseRecorder := httptest.NewRecorder()
		//testContext, _ := gin.CreateTestContext(responseRecorder)

		t.Run("When controller.ProviderRegister return nil error, should return success resp", func(t *testing.T) {
			//patch := gomonkey.ApplyMethodFunc(
			//	reflect.TypeOf(mockAPI.controller),
			//	"ProviderRegister",
			//	func(_ goContext.Context, _ []byte) error {
			//		return nil
			//	},
			//)
			//defer patch.Reset()
			//
			//req, err := http.NewRequest("POST", "http://127.0.0.1", bytes.NewBufferString("test body"))
			//testContext.Request = req
			//if err != nil {
			//	t.Error(err)
			//}
			//mockAPI.providerRegister(testContext)
			//respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)
			//if err != nil {
			//	t.Error(err)
			//}
			//
			//var resp types.ResponseJson
			//if err = json.Unmarshal(respBody, &resp); err != nil {
			//	t.Error(err)
			//}
			//
			//So(resp.Code, ShouldEqual, http.StatusOK)
			//So(resp.Message, ShouldEqual, "register success")
		})

		t.Run("When controller.ProviderRegister return an error, should return an error resp", func(t *testing.T) {
			//patch := gomonkey.ApplyMethodFunc(
			//	reflect.TypeOf(mockAPI.controller),
			//	"ProviderRegister",
			//	func(_ goContext.Context, _ []byte) error {
			//		return fmt.Errorf("monkey error")
			//	},
			//)
			//defer patch.Reset()
			//
			//req, err := http.NewRequest("POST", "http://127.0.0.1", bytes.NewBufferString("test body"))
			//testContext.Request = req
			//if err != nil {
			//	t.Error(err)
			//}
			//
			//mockAPI.providerRegister(testContext)
			//respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)
			//
			//var resp types.ResponseJson
			//if err = json.Unmarshal(respBody, &resp); err != nil {
			//	t.Error(err)
			//}
			//
			//So(resp.Code, ShouldEqual, http.StatusBadRequest)
			//So(resp.Message, ShouldEqual, "monkey error")
		})
	})
}
