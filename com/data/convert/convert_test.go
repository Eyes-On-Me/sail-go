package convert_test

import (
	"net/http"
	"github.com/sail-services/sail-go/com/data/convert"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConvert(t *testing.T) {
	Convey("string 到 float64", t, func() {
		So(convert.SToF64("21312421.213123"), ShouldEqual, 2.1312421213123e+07)
	})
	Convey("float64 到 string", t, func() {
		So(convert.F64ToS(21312421.213123), ShouldEqual, "21312421.213123")
	})
	Convey("string 到 int64", t, func() {
		So(convert.SToI64("123"), ShouldEqual, 123)
	})
	Convey("int64 到 string", t, func() {
		So(convert.I64ToS(1000000000023), ShouldEqual, "1000000000023")
	})
	Convey("int 到 string", t, func() {
		So(convert.IToS(123), ShouldEqual, "123")
	})
	Convey("*http.Response 到 []byte", t, func() {
		resp, _ := http.Get("http://www.baidu.com")
		So(convert.RespToB(resp), ShouldNotBeEmpty)
	})
	Convey("Hex 到 int", t, func() {
		hex := map[string]int{
			"1":   1,
			"002": 2,
			"011": 17,
			"0a1": 161,
			"35e": 862,
		}
		for h, dec := range hex {
			val, err := convert.HexSToI(h)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, dec)
		}
	})
	Convey("int 到 Hex", t, func() {
		dec := map[int]string{
			1:   "1",
			2:   "2",
			17:  "11",
			161: "a1",
			862: "35e",
		}
		for d, hex := range dec {
			val := convert.IToHexS(d)
			So(val, ShouldEqual, hex)
		}
	})
}
