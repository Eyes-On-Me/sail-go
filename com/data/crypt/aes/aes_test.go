package aes_test

import (
	"encoding/base64"
	"github.com/sail-services/sail-go/foundation/component/data/crypt/aes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAES(t *testing.T) {
	Convey("AES 加密 / 解密", t, func() {
		key := []byte("sfe023f_9fd&fwfl")
		es, e := aes.Encrypt([]byte("hello"), key)
		So(e, ShouldBeNil)
		So(base64.StdEncoding.EncodeToString(es), ShouldEqual, "R02+VNvEI59gLMyDXZ9DdQ==")
		ds, e := aes.Decrypt(es, key)
		So(e, ShouldBeNil)
		So(string(ds), ShouldEqual, "hello")
	})
}
