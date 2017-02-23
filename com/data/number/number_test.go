package number_test

import (
	"math"
	"github.com/sail-services/sail-go/com/data/number"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNumber(t *testing.T) {
	Convey("四舍五入", t, func() {
		So(number.Round(math.Pi, 2), ShouldEqual, 3.14)
	})
}
