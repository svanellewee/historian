package storage

import (
	"bytes"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// History structure
type History struct {
	id        int64
	historyID int64
	data      string
	time      time.Time
}

func (h History) String() string {
	return fmt.Sprintf("[%s] %s", h.time.Format(time.RFC3339), h.data)
}

// Store bolddb structure
type Store struct {
	db *bolt.DB
}

// Close on stores
func (s *Store) Close() {
	s.db.Close()
}

// Dump stores
func (s *Store) Dump(bucket string) {
	fmt.Println("Dump........", bucket)
	s.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				snapshotTime, err := StringToTime(string(k))
				if err != nil {
					return err
				}
				fmt.Printf("key=%s, value=%s\n", snapshotTime.Format(time.RFC3339), string(v))
				return nil
			})
		}
		return nil
	})
}

// NewStore to create a storage file
func NewStore(dbFile string) (*Store, error) {
	db, err := bolt.Open(dbFile, 0777, nil)
	if err != nil {
		logrus.Errorf("could not open (%s) [%v]", dbFile, err)
		return nil, err
	}
	return &Store{
		db: db,
	}, nil
}

// Add to storage
func (s *Store) Add(currentDirectory string, time time.Time, commandEntry []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(currentDirectory))
		if err != nil {
			return err
		}
		ts := []byte(TimeToString(time))
		err = b.Put(ts, commandEntry)
		if err != nil {
			return err
		}
		return nil
	})
}

// Get from storage
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

// Range over storage between dates
func (s *Store) Range(bucket string, minTime, maxTime time.Time, handler func(t time.Time, data []byte)) {
	s.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket([]byte(bucket))
		c := mainBucket.Cursor()
		min := []byte(TimeToString(minTime))
		max := []byte(TimeToString(maxTime))
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			tt, err := StringToTime(string(k))
			if err != nil {
				return err
			}
			handler(tt, v)
		}
		return nil
	})
}

type bucketHandler func(name []byte, b *bolt.Bucket) error

// ForEachBucket apply a specified handler function
func (s *Store) ForEachBucket(handleBucket bucketHandler) error {
	return s.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		err := tx.ForEach(handleBucket)
		if err != nil {
			return err
		}
		return nil
	})
}

type bucketKeyValueHandler func(name []byte, bucket *bolt.Bucket, key []byte, value []byte) error

func oneBucketForDay(name []byte, bucket *bolt.Bucket, timestamp time.Time, handleKeyValue bucketKeyValueHandler) error {
	c := bucket.Cursor()
	prefix := makePrefixKeyDate(timestamp)
	for key, value := c.Seek(prefix); key != nil && bytes.HasPrefix(key, prefix); key, value = c.Next() {
		err := handleKeyValue(name, bucket, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// AllBucketsForDay something something...also does a today function
func (s *Store) AllBucketsForDay(requestedTime time.Time, handler bucketKeyValueHandler) error {
	return s.ForEachBucket(func(name []byte, b *bolt.Bucket) error {
		return oneBucketForDay(name, b, requestedTime, handler)
	})
}

func makePrefixKeyDate(timestamp time.Time) []byte {
	year, month, day := timestamp.Date()
	bod := time.Date(year, month, day, 0, 0, 0, 0, &time.Location{})
	return []byte(TimeToString(bod)[:10])
}

// Today gets the bucket entries for specified date
func (s *Store) Today(bucket string, prefixTime time.Time, handler func(string, []byte)) {
	s.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket([]byte(bucket))
		c := mainBucket.Cursor()
		prefix := makePrefixKeyDate(prefixTime)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			handler(string(prefix), v)
		}
		return nil
	})
}

// Last n entries
func (s *Store) Last(directory string, numEntries int) ([]History, error) {
	historyList := make([]History, 0, numEntries)
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
			historyValue := History{
				data: string(v),
				time: timeValue,
			}
			historyList = append(historyList, historyValue)
			i--
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return historyList, nil
}

// TimeToString converts time to string
func TimeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

// StringToTime converts string to time
func StringToTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
