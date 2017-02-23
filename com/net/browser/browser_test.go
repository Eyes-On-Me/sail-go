package browser_test

import (
	"github.com/sail-services/sail-go/com/network/browser"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBrowser(t *testing.T) {
	Convey("是否为移动端", t, func() {
		So(browser.IsMobile("Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/35.0.1916.153 Safari/537.36"), ShouldBeFalse)
	})
}
