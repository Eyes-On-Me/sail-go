package uuid

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"net"
	"os"
	"sync"
	"time"
)

type UUID [16]byte

const (
	VARIANT_NCS = iota
	VARIANT_RFC4122
	VARIANT_MICROSOFT
	VARIANT_FUTURE
	DOMAIN_PERSON = iota
	DOMAIN_GROUP
	DOMAIN_ORG
	dash        byte = '-'
	epoch_start      = 122192928000000000
)

var (
	epoch_func        func() uint64
	storage_mutex     sync.Mutex
	clock_sequence    uint16
	last_time         uint64
	hardware_addr     [6]byte
	posix_uid         = uint32(os.Getuid())
	posix_gid         = uint32(os.Getgid())
	urn_prefix        = []byte("urn:uuid:")
	byte_groups       = []int{8, 4, 4, 4, 12}
	NIL               = UUID{}
	NAMESPACE_DNS, _  = FromS("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	NAMESPACE_URL, _  = FromS("6ba7b811-9dad-11d1-80b4-00c04fd430c8")
	NAMESPACE_OID, _  = FromS("6ba7b812-9dad-11d1-80b4-00c04fd430c8")
	NAMESPACE_X500, _ = FromS("6ba7b814-9dad-11d1-80b4-00c04fd430c8")
)

func init() {
	buf := make([]byte, 2)
	rand.Read(buf)
	clock_sequence = binary.BigEndian.Uint16(buf)
	rand.Read(hardware_addr[:])
	hardware_addr[0] |= 0x01
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if len(iface.HardwareAddr) >= 6 {
				copy(hardware_addr[:], iface.HardwareAddr)
				break
			}
		}
	}
	epoch_func = unixTimeFunc
}

// New_V1 returns UUID based on current timestamp and MAC address.
func NewV1() UUID {
	u := UUID{}
	timeNow, clockSeq := getStorage()
	binary.BigEndian.PutUint32(u[0:], uint32(timeNow))
	binary.BigEndian.PutUint16(u[4:], uint16(timeNow>>32))
	binary.BigEndian.PutUint16(u[6:], uint16(timeNow>>48))
	binary.BigEndian.PutUint16(u[8:], clockSeq)
	copy(u[10:], hardware_addr[:])
	u.VersionSet(1)
	u.VariantSet()
	return u
}

// New_V2 returns DCE Security UUID based on POSIX UID/GID.
func NewV2(domain byte) UUID {
	u := UUID{}
	switch domain {
	case DOMAIN_PERSON:
		binary.BigEndian.PutUint32(u[0:], posix_uid)
	case DOMAIN_GROUP:
		binary.BigEndian.PutUint32(u[0:], posix_gid)
	}
	timeNow, clockSeq := getStorage()
	binary.BigEndian.PutUint16(u[4:], uint16(timeNow>>32))
	binary.BigEndian.PutUint16(u[6:], uint16(timeNow>>48))
	binary.BigEndian.PutUint16(u[8:], clockSeq)
	u[9] = domain
	copy(u[10:], hardware_addr[:])
	u.VersionSet(2)
	u.VariantSet()
	return u
}

// New_V3 returns UUID based on MD5 hash of namespace UUID and name.
func NewV3(ns UUID, name string) UUID {
	u := newFromHash(md5.New(), ns, name)
	u.VersionSet(3)
	u.VariantSet()
	return u
}

// New_V4 returns random generated UUID.
func NewV4() UUID {
	u := UUID{}
	rand.Read(u[:])
	u.VersionSet(4)
	u.VariantSet()
	return u
}

// New_V5 returns UUID based on SHA-1 hash of namespace UUID and name.
func NewV5(ns UUID, name string) UUID {
	u := newFromHash(sha1.New(), ns, name)
	u.VersionSet(5)
	u.VariantSet()
	return u
}

func And(u1 UUID, u2 UUID) UUID {
	u := UUID{}
	for i := 0; i < 16; i++ {
		u[i] = u1[i] & u2[i]
	}
	return u
}

func Or(u1 UUID, u2 UUID) UUID {
	u := UUID{}
	for i := 0; i < 16; i++ {
		u[i] = u1[i] | u2[i]
	}
	return u
}

func Equal(u1 UUID, u2 UUID) bool {
	return bytes.Equal(u1[:], u2[:])
}

func (u UUID) Version() uint {
	return uint(u[6] >> 4)
}

func (u UUID) Variant() uint {
	switch {
	case (u[8] & 0x80) == 0x00:
		return VARIANT_NCS
	case (u[8]&0xc0)|0x80 == 0x80:
		return VARIANT_RFC4122
	case (u[8]&0xe0)|0xc0 == 0xc0:
		return VARIANT_MICROSOFT
	}
	return VARIANT_FUTURE
}

func (u UUID) B() []byte {
	return u[:]
}

// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func (u UUID) S() string {
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], u[0:4])
	buf[8] = dash
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = dash
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = dash
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = dash
	hex.Encode(buf[24:], u[10:])
	return string(buf)
}

func (u *UUID) VersionSet(v byte) {
	u[6] = (u[6] & 0x0f) | (v << 4)
}

func (u *UUID) VariantSet() {
	u[8] = (u[8] & 0xbf) | 0x80
}

func (u UUID) TextMarshal() (text []byte, err error) {
	text = []byte(u.S())
	return
}

// Following formats are supported:
// "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
// "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}"
// "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"
func (u *UUID) TextUnmarshal(text []byte) (err error) {
	if len(text) < 32 {
		err = fmt.Errorf("uuid: invalid UUID string: %s", text)
		return
	}
	if bytes.Equal(text[:9], urn_prefix) {
		text = text[9:]
	} else if text[0] == '{' {
		text = text[1:]
	}
	b := u[:]
	for _, byteGroup := range byte_groups {
		if text[0] == '-' {
			text = text[1:]
		}
		_, err = hex.Decode(b[:byteGroup/2], text[:byteGroup])
		if err != nil {
			return
		}
		text = text[byteGroup:]
		b = b[byteGroup/2:]
	}
	return
}

func (u UUID) BinaryMarshal() (data []byte, err error) {
	data = u.B()
	return
}

func (u *UUID) BinaryUnmarshal(data []byte) (err error) {
	if len(data) != 16 {
		err = fmt.Errorf("uuid: UUID must be exactly 16 bytes long, got %d bytes", len(data))
		return
	}
	copy(u[:], data)
	return
}

func (u *UUID) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		if len(src) == 16 {
			return u.BinaryUnmarshal(src)
		}
		return u.TextUnmarshal(src)
	case string:
		return u.TextUnmarshal([]byte(src))
	}
	return fmt.Errorf("uuid: cannot convert %T to UUID", src)
}

func FromB(input []byte) (u UUID, err error) {
	err = u.BinaryUnmarshal(input)
	return
}

func FromS(input string) (u UUID, err error) {
	err = u.TextUnmarshal([]byte(input))
	return
}

func unixTimeFunc() uint64 {
	return epoch_start + uint64(time.Now().UnixNano()/100)
}

func getStorage() (uint64, uint16) {
	storage_mutex.Lock()
	defer storage_mutex.Unlock()
	timeNow := epoch_func()
	if timeNow <= last_time {
		clock_sequence++
	}
	last_time = timeNow
	return timeNow, clock_sequence
}

func newFromHash(h hash.Hash, ns UUID, name string) UUID {
	u := UUID{}
	h.Write(ns[:])
	h.Write([]byte(name))
	copy(u[:], h.Sum(nil))
	return u
}
