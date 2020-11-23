package storage_test

import (
	"fmt"
	"os"
	"regexp"
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

func TestConvertHelper(t *testing.T) {
	testCases := []struct {
		directory  string
		timestamp  time.Time
		command    string
		result     string
		annotation string
		id         int64
	}{
		{
			directory:  "/tmp",
			timestamp:  time.Date(2020, 1, 1, 0, 2, 0, 0, &time.Location{}),
			command:    "1234 ls /hello",
			id:         1234,
			result:     "ls /hello",
			annotation: "",
		},
	}

	for _, testCase := range testCases {

		_, err := storage.Convert(testCase.command)
		assert.Nil(t, err)

		//		assert.EqualValues(t, testCase.result, history.command)
	}
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
		history, err := storage.NewHistory(
			testCase.command,
			storage.SetDirectory(testCase.directory),
			storage.SetTime(testCase.timestamp),
		)
		assert.Nil(t, err)
		err = store.Add(
			history,
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
	testCases := []struct {
		directory string
		timestamp time.Time
		command   string
	}{
		{
			directory: "/tmp",
			timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, &time.Location{}),
			command:   "ls .",
		},
		{
			directory: "/tmp",
			timestamp: time.Date(2020, 1, 1, 23, 55, 55, 55, &time.Location{}),
			command:   "source ./bla",
		},
		{
			directory: "/tmp",
			timestamp: time.Date(2020, 1, 2, 1, 0, 2, 0, &time.Location{}),
			command:   "echo bla",
		},
		{
			directory: "/home/user",
			timestamp: time.Date(2020, 1, 1, 1, 1, 2, 0, &time.Location{}),
			command:   "echo bla2",
		},
		{
			directory: "/home/user2",
			timestamp: time.Date(2020, 1, 2, 1, 2, 2, 0, &time.Location{}),
			command: `cat<<"EOF" > test
			bla yadda
			EOF
			`,
		},
	}

	for _, testCase := range testCases {
		history, err := storage.NewHistory(
			testCase.command,
			storage.SetDirectory(testCase.directory),
			storage.SetTime(testCase.timestamp),
		)
		assert.Nil(t, err)
		err = store.Add(history)
		assert.Nil(t, err)
	}

	//store.Dump("/tmp")
	//store.Dump("/home/user2")
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

	filter := func(directoryName []byte, key []byte, value []byte) bool {
		re, err := regexp.Compile("echo")
		if err != nil {
			return false
		}
		return re.Find(value) != nil
	}
	fmt.Println("search/filter")

	results, err := store.All() // maybe we need an store.All function
	assert.Nil(t, err)
	assert.Equal(t, 5, len(results))

	filterResults, err := store.All(filter)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(filterResults))
	for _, filterResult := range filterResults {
		fmt.Println(">>", filterResult)
	}

}
