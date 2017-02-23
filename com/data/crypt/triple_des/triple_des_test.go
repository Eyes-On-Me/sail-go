package triple_des_test

import (
	"github.com/sail-services/sail-go/com/data/crypt/base64"
	"github.com/sail-services/sail-go/com/data/crypt/triple_des"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTripleDES(t *testing.T) {
	Convey("3DES 加密 / 解密", t, func() {
		key := []byte("sfe023f_sefiel#fi32lf3e!")
		es, e := triple_des.Encrypt([]byte("hello"), key)
		So(e, ShouldBeNil)
		So(base64.Encode(es), ShouldEqual, "1vULm9iU624=")
		ds, e := triple_des.Decrypt(es, key)
		So(e, ShouldBeNil)
		So(string(ds), ShouldEqual, "hello")
	})
}
