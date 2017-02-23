package random_test

import (
	"github.com/sail-services/sail-go/com/data/random"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRandom(t *testing.T) {
	Convey("随机数字", t, func() {
		So(random.I(0, 9), ShouldNotEqual, 10)
	})
	Convey("随机字符串", t, func() {
		So(random.S(random.RANDOM_STRING, 4), ShouldNotBeNil)
	})
}
