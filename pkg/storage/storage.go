package storage

import (
	"bytes"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// History structure
type History struct {
	id        int64
	historyID int64
	data      string
	time      time.Time
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
				snapshotTime, err := stringToTime(string(k))
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
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
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
		ts := []byte(timeToString(time))
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

// Today gets the bucket entries for specified date
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

func timeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

func stringToTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
