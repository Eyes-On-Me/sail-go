package fs_test

import (
	"github.com/sail-services/sail-go/com/sys/fs"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFS(t *testing.T) {
	var f string
	var e error
	Convey("生成目录", t, func() {
		e = fs.PathNew("test", 0755)
		So(e, ShouldBeNil)
		e = fs.Delete("test")
		So(e, ShouldBeNil)
	})
	Convey("写文件", t, func() {
		content := "teststttttt"
		l, e := fs.FilePutS(f+"1.txt", content)
		So(e, ShouldBeNil)
		So(l, ShouldNotBeNil)
	})
	Convey("删除文件", t, func() {
		e = fs.Delete("1.txt")
		So(e, ShouldBeNil)
	})
}
