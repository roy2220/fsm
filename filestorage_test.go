package fsm_test

import (
	"encoding/binary"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/roy2220/fsm"
	"github.com/stretchr/testify/assert"
)

const N = 1000000

type Entry struct {
	KeySize uint8
	KeyHash uint64
	KeyPtr  int64
}

var Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func TestFileStorage(t *testing.T) {
	const fn = "./test_storage"
	defer func() { t.Log(os.Remove(fn)) }()
	Store(t, fn)
	Load(t, fn)
}

func Store(t *testing.T, fn string) {
	fs := new(fsm.FileStorage).Init()
	err := fs.Open(fn, true)

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	defer func() {
		fs.Close()
		st := fs.Stats()
		t.Logf("stats: %#v", st)
	}()

	es := make([]Entry, N)

	for i := range es {
		es[i] = MakeEntry(fs)

		if i%2 == 1 {
			j := rand.Intn(i)
			fs.FreeSpace(es[j].KeyPtr)
			es[j] = MakeEntry(fs)
		}
	}

	s, buf := fs.AllocateSpace(9*N + 8)
	fs.SetPrimarySpace(s)
	checksum := uint64(0)

	for i := range es {
		j := i * 9
		binary.BigEndian.PutUint64(buf[j:], uint64(es[i].KeyPtr))
		buf[j+8] = es[i].KeySize
		checksum ^= es[i].KeyHash
	}

	binary.BigEndian.PutUint64(buf[9*N:], checksum)
}

func Load(t *testing.T, fn string) {
	fs := new(fsm.FileStorage).Init()
	err := fs.Open(fn, true)

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	defer func() {
		fs.Close()
		st := fs.Stats()
		t.Logf("stats: %#v", st)
		assert.Equal(t, 0, st.UsedSpaceSize)
	}()

	buf := fs.AccessSpace(fs.PrimarySpace())
	checksum := uint64(0)

	for i := 0; i < N; i++ {
		j := i * 9
		kp := int64(binary.BigEndian.Uint64(buf[j:]))
		buf2 := fs.AccessSpace(kp)
		ks := int(buf[j+8])
		k := buf2[:ks]
		kh := HashKey(k)
		checksum ^= kh
		fs.FreeSpace(kp)
	}

	checksum2 := uint64(binary.BigEndian.Uint64(buf[9*N:]))
	assert.Equal(t, checksum, checksum2)
	fs.FreeSpace(fs.PrimarySpace())
}

func MakeEntry(fs *fsm.FileStorage) Entry {
	k := GenerateKey()
	kp, buf := fs.AllocateSpace(len(k))
	copy(buf, k)

	return Entry{
		KeySize: uint8(len(k)),
		KeyHash: HashKey(k),
		KeyPtr:  kp,
	}
}

func GenerateKey() []byte {
	ks := Rand.Intn(256)
	k := make([]byte, ks)

	for i := range k {
		k[i] = byte(Rand.Intn(256))
	}

	return k
}

func HashKey(k []byte) uint64 {
	kh := uint64(0)

	for i := range k {
		kh ^= kh*131 + uint64(k[i])
	}

	return kh
}
