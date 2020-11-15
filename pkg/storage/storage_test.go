package storage_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/svanellewee/historian/pkg/storage"
	bolt "go.etcd.io/bbolt"
)

func TestTimeConvert(t *testing.T) {
	timestamp1 := time.Date(2020, 1, 1, 0, 2, 0, 0, &time.Location{})
	s := storage.TimeToString(timestamp1)
	tme, err := storage.StringToTime(s)
	assert.Nil(t, err)
	fmt.Printf("%#v %s", tme, tme)
}

func TestBucketList(t *testing.T) {
	dbFile := "my.db"

	store, err := storage.NewStore(dbFile)
	assert.Nil(t, err)
	defer os.Remove(dbFile)

	testCases := []struct {
		directory string
		timestamp time.Time
		command   string
	}{
		{
			directory: "/tmp",
			timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, &time.Location{}),
			command:   "ls",
		},
		{
			directory: "/tmp",
			timestamp: time.Date(2020, 1, 1, 2, 0, 0, 0, &time.Location{}),
			command:   "echo something",
		},
		{
			directory: "/tmp",
			timestamp: time.Date(2020, 1, 2, 2, 0, 0, 0, &time.Location{}),
			command:   "echo something else",
		},
		{
			directory: "/tmp2",
			timestamp: time.Date(2020, 1, 2, 0, 0, 0, 0, &time.Location{}),
			command:   "ls",
		},
		{
			directory: "/tmp2",
			timestamp: time.Date(2020, 1, 1, 2, 0, 0, 0, &time.Location{}),
			command:   "echo something",
		},
		{
			directory: "/tmp2",
			timestamp: time.Date(2020, 1, 2, 2, 0, 0, 0, &time.Location{}),
			command:   "echo something else",
		},
	}

	for _, testCase := range testCases {
		err = store.Add(
			testCase.directory,
			testCase.timestamp,
			[]byte(testCase.command),
		)
		assert.Nil(t, err)
	}

	expectedDirectories := 2 // tmp tmp2
	actualDirectories := 0
	store.ForEachBucket(func(name []byte, _ *bolt.Bucket) error {
		fmt.Println(string(name))
		actualDirectories++
		return nil
	})
	assert.Equal(t, expectedDirectories, actualDirectories)
}

func TestAnother(t *testing.T) {
	dbFile := "my.db"

	store, err := storage.NewStore(dbFile)
	assert.Nil(t, err)
	defer os.Remove(dbFile)
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

	err = store.AllBucketsForDay(minTime, func(name []byte, b *bolt.Bucket, k []byte, v []byte) error {
		fmt.Printf("All Buckets For Day [%s] key=%s, value=%s\n", string(name), k, v)
		return nil
	})
	assert.Nil(t, err)

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
