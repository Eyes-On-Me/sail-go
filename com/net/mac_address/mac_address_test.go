package mac_address_test

import (
	"github.com/sail-services/sail-go/com/net/mac_address"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMacAddress(t *testing.T) {
	Convey("生成一个随机 MAC 地址", t, func() {
		mac := mac_address.Random(":")
		So(mac, ShouldNotBeEmpty)
	})
}
