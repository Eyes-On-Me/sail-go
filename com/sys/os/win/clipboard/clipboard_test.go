package clipboard_test

import (
	"github.com/sail-services/sail-go/com/sys/os/win/clipboard"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCliboard(t *testing.T) {
	Convey("剪切板", t, func() {
		clipboard.Set("a")
		So(clipboard.Get(), ShouldEqual, "a")
	})
}
