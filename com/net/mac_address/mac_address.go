package mac_address

import (
	"github.com/sail-services/sail-go/com/data/random"
	"fmt"
	"strings"
)

func Random(split string) (ret string) {
	cid := [][]string{
		{"90", "2b", "34"},
		{"f4", "ec", "38"},
		{"00", "50", "56"},
		{"00", "16", "3e"},
		{"52", "54", "00"}}
	mac := []string{}
	mac = append(mac, cid[random.I(0, 4)]...)
	mac = append(mac, fmt.Sprintf("%02x", random.I(0x00, 0x7f)))
	mac = append(mac, fmt.Sprintf("%02x", random.I(0x00, 0xff)))
	mac = append(mac, fmt.Sprintf("%02x", random.I(0x00, 0xfe)))
	for i, s := range mac {
		mac[i] = strings.ToUpper(s)
	}
	ret = strings.Join(mac, split)
	return
}
