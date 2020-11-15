package storage_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

type history struct {
	id        int64
	historyID int64
	data      string
	time      time.Time
}

type histories []history
type directoryName string
type HistoryStore map[directoryName]history

type Store struct {
	db *bolt.DB
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Dump(bucket string) {
	fmt.Println("Dump........", bucket)
	s.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				snapshotTime, err := stringToTime(string(k))
				if err != nil {
					return err
				}
				fmt.Println(">>>", k)
				fmt.Printf("key=%s, value=%s\n", snapshotTime.Format(time.RFC3339), v)
				return nil
			})
		}
		return nil
	})
}

func NewStore(dbFile string) (*Store, error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &Store{
		db: db,
	}, nil
}

func (s *Store) Add(currentDirectory string, time time.Time, commandEntry []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(currentDirectory))
		if err != nil {
			return err
		}
		ts := []byte(timeToString(time))
		err = b.Put(ts, commandEntry)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) Get(bucket, key string) ([]byte, error) {
	var result []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		result = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) Range(bucket string, minTime, maxTime time.Time, handler func(t time.Time, data []byte)) {
	s.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket([]byte(bucket))
		c := mainBucket.Cursor()
		min := []byte(timeToString(minTime))
		max := []byte(timeToString(maxTime))
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			tt, err := stringToTime(string(k))
			if err != nil {
				return err
			}
			handler(tt, v)
		}
		return nil
	})
}

func (s *Store) Today(bucket string, prefixTime time.Time, handler func(string, []byte)) {
	s.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket([]byte(bucket))
		c := mainBucket.Cursor()
		year, month, day := prefixTime.Date()
		bod := time.Date(year, month, day, 0, 0, 0, 0, &time.Location{})
		prefix := timeToString(bod)[:10]
		for k, v := c.Seek([]byte(prefix)); k != nil && bytes.HasPrefix(k, []byte(prefix)); k, v = c.Next() {
			handler(prefix, v)
		}
		return nil
	})
}

func (s *Store) Last(directory string, numEntries int) ([]history, error) {
	historyList := make([]history, 0, numEntries)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(directory))
		c := b.Cursor()
		i := numEntries
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if i <= 0 {
				break
			}
			timeValue, err := time.Parse(time.RFC3339, string(k))
			if err != nil {
				return err
			}
			historyValue := history{
				data: string(v),
				time: timeValue,
			}
			historyList = append(historyList, historyValue)
			i -= 1
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return historyList, nil
}

func timeToBytes(t time.Time) []byte {
	nanoTime := uint64(t.UnixNano())
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, nanoTime)
	return result
}

func bytesToTime(timeBytes []byte) time.Time {
	return time.Unix(0, int64(binary.BigEndian.Uint64(timeBytes)))
}

func TestTimeConvert(t *testing.T) {
	timestamp1 := time.Date(2020, 1, 1, 0, 0, 0, 0, &time.Location{})
	d := timeToBytes(timestamp1)
	revertD := bytesToTime(d)
	fmt.Println(revertD.Format(time.RFC3339))
}

func timeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

func stringToTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
func TestTimeConvert2(t *testing.T) {
	timestamp1 := time.Date(2020, 1, 1, 0, 2, 0, 0, &time.Location{})
	s := timeToString(timestamp1)
	tme, err := stringToTime(s)
	assert.Nil(t, err)
	fmt.Printf("%#v %s", tme, tme)
}
func TestAnother(t *testing.T) {
	dbFile := "my.db"
	os.Remove(dbFile)

	store, err := NewStore(dbFile)
	assert.Nil(t, err)

	timestamp1 := time.Date(2020, 1, 1, 0, 0, 0, 0, &time.Location{})
	fmt.Println(timestamp1.Format(time.RFC3339), "......!")
	err = store.Add(
		"/tmp",
		timestamp1,
		[]byte("ls ."),
	)
	assert.Nil(t, err)

	timestamp1a := time.Date(2020, 1, 1, 23, 55, 55, 55, &time.Location{})
	fmt.Println(timestamp1.Format(time.RFC3339), "......!")
	err = store.Add(
		"/tmp",
		timestamp1a,
		[]byte("source ./bla"),
	)
	assert.Nil(t, err)

	timestamp2 := time.Date(2020, 1, 2, 1, 0, 2, 0, &time.Location{})
	err = store.Add(
		"/tmp",
		timestamp2,
		[]byte("echo bla"),
	)
	assert.Nil(t, err)

	timestamp3 := time.Date(2020, 1, 1, 1, 1, 2, 0, &time.Location{})
	err = store.Add(
		"/home/user",
		timestamp3,
		[]byte("echo bla"),
	)
	assert.Nil(t, err)

	timestamp4 := time.Date(2020, 1, 2, 1, 2, 2, 0, &time.Location{})
	err = store.Add(
		"/home/user2",
		timestamp4,
		[]byte(`cat<<"EOF" > test
bla yadda
EOF
`),
	)
	assert.Nil(t, err)

	store.Dump("/tmp")
	store.Dump("/home/user2")
	minTime := time.Date(2020, 1, 1, 0, 0, 0, 0, &time.Location{})
	maxTime := time.Date(2020, 1, 3, 0, 0, 2, 0, &time.Location{})

	fmt.Println("\nRange test..")
	rangeCount := 0
	store.Range("/tmp", minTime, maxTime, func(t time.Time, value []byte) {
		rangeCount += 1
		fmt.Printf("RANGE  [%s]: %s\n", t.Format(time.RFC3339), value)
	})
	assert.Equal(t, 3, rangeCount)

	fmt.Println("\nPrefix test")
	prefixCount := 0
	store.Today("/tmp", minTime, func(k string, v []byte) {
		prefixCount += 1
		fmt.Printf("PREFIX key=%s, value=%s\n", k, v)
	})
	assert.Equal(t, 2, prefixCount)

	fmt.Println("Last...")
	lastEntries, err := store.Last("/tmp", 1)
	assert.Nil(t, err)
	assert.Len(t, lastEntries, 1)

	fmt.Println("Last...2")
	lastEntries, err = store.Last("/tmp", 2)
	assert.Nil(t, err)
	assert.Len(t, lastEntries, 2)

	fmt.Println("Last...4")
	lastEntries, err = store.Last("/tmp", 4)
	fmt.Println(lastEntries)
	assert.Nil(t, err)
	assert.Len(t, lastEntries, 3)
}
