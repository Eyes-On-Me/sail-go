package base64_test

import (
	"github.com/sail-services/sail-go/com/data/crypt/base64"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBase64(t *testing.T) {
	Convey("Base64 编码 / 解码", t, func() {
		e := base64.EncodeS("hello")
		So(e, ShouldEqual, "aGVsbG8=")
		d, err := base64.Decode(e)
		So(err, ShouldBeNil)
		So(d, ShouldEqual, "hello")
	})
}
