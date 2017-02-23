package des_test

import (
	"encoding/base64"
	"github.com/sail-services/sail-go/com/data/crypt/des"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDES(t *testing.T) {
	Convey("DES 加密 / 解密", t, func() {
		key := []byte("sfe023f_")
		es, e := des.Encrypt([]byte("hello"), key)
		So(e, ShouldBeNil)
		So(base64.StdEncoding.EncodeToString(es), ShouldEqual, "bvn6fko5Hk0=")
		ds, e := des.Decrypt(es, key)
		So(e, ShouldBeNil)
		So(string(ds), ShouldEqual, "hello")
	})
}
