package md5_test

import (
	"github.com/sail-services/sail-go/com/data/crypt/md5"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMD5(t *testing.T) {
	Convey("MD5 编码",t, func() {
		So(md5.Encode("hello"), ShouldEqual, "5d41402abc4b2a76b9719d911017c592")
	})
}
