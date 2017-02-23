package download_test

import (
	"github.com/sail-services/sail-go/com/network/download"
	"github.com/sail-services/sail-go/com/system/fs"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDownload(t *testing.T) {
	Convey("下载文件", t, func() {
		e := download.Download("http://static.youku.com/index/img/header/yklogo.png", "test.png")
		So(e, ShouldBeNil)
		e = fs.Delete("test.png")
		So(e, ShouldBeNil)
	})
}
