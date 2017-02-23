package fetcher_test

import (
	"net/url"
	"github.com/sail-services/sail-go/mod/net/fetcher"
	"testing"

	"github.com/PuerkitoBio/goquery"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_Fetch(t *testing.T) {
	Convey("测试 GET", t, func() {
		f := fetcher.New("baidu.com")
		resp, err := f.Get("/")
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
	})
	Convey("测试 POST", t, func() {
		f := fetcher.New("alibench.com")
		f.Get("/")
		data := url.Values{
			"task_from": {"self"},
			"target":    {"http://golang.org"},
			"ac":        {"http"},
		}
		resp, err := f.PostForm("/new_task.php", data)
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
	})
	Convey("获取网页源码 GoQuery 分析源码", t, func() {
		f := fetcher.New("www.baidu.com")
		resp, err := f.Get("/")
		So(err, ShouldBeNil)
		doc, err := goquery.NewDocumentFromResponse(resp)
		So(err, ShouldBeNil)
		txt := doc.Find("title").Text()
		So(txt, ShouldEqual, "百度一下，你就知道")
	})
}
