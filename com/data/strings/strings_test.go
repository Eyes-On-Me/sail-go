package strings_test

import (
	"github.com/sail-services/sail-go/com/data/strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStrings(t *testing.T) {
	Convey("string 居中", t, func() {
		s := strings.Center("aaa", 5)
		So(s, ShouldEqual, " aaa ")
	})
	Convey("string 反转", t, func() {
		s := strings.Reverse("abc")
		So(s, ShouldEqual, "cba")
	})
}
