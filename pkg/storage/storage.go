package storage

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// History structure
type History struct {
	ID            int64
	historyID     int64
	Data          string
	Time          time.Time
	DirectoryName string
	Annotation    string
}

// HistOption updates History structs.
type HistOption func(h *History) error

func SetTime(t time.Time) HistOption {
	return func(h *History) error {
		h.Time = t
		return nil
	}
}

func SetDirectory(directory string) HistOption {
	return func(h *History) error {
		h.DirectoryName = directory
		return nil
	}
}

func SetID(id int64) HistOption {
	return func(h *History) error {
		h.ID = id
		return nil
	}
}

func SetAnnotation(annotation string) HistOption {
	return func(h *History) error {
		h.Annotation = annotation
		return nil
	}
}

// NewHistory returns a new history entry
func NewHistory(command string, options ...HistOption) (*History, error) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	history := &History{
		Time:          time.Now(), // default to current time.
		DirectoryName: currentDirectory,
		Data:          command,
	}

	for _, option := range options {
		err := option(history)
		if err != nil {
			return nil, err
		}
	}
	return history, nil
}

// ErrIncorrectCount reports unexpected argument counts.
type ErrIncorrectCount struct {
	ArgumentCount int
	ExpectedCount int
}

func (e ErrIncorrectCount) Error() string {
	return fmt.Sprintf("incorrect number of elements, Expected [%d], found [%d]", e.ArgumentCount, e.ExpectedCount)
}

// Convert strings of form "1234 ls /tmp # some annotation" to a history entry.
func Convert(input string) (*History, error) {
	// This is buggy :-/
	trimmed := strings.Trim(input, " ")
	elements := strings.SplitN(trimmed, " ", 2)
	if len(elements) != 2 {
		return nil, ErrIncorrectCount{ArgumentCount: len(elements), ExpectedCount: 2}
	}

	number, err := strconv.ParseInt(elements[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse error %w", err)
	}

	options := []HistOption{
		SetID(number),
	}

	history, err := NewHistory(elements[1], options...)
	if err != nil {
		return nil, err
	}

	return history, nil
}

// _ConvertAnnotate is a broken implementation of convert that does not properly separate comments from commands. Better plan required
func _ConvertAnnotate(input string) (*History, error) {
	// This is buggy :-/
	trimmed := strings.Trim(input, " ")
	elements := strings.SplitN(trimmed, " ", 2)
	if len(elements) != 2 {
		return nil, ErrIncorrectCount{ArgumentCount: len(elements), ExpectedCount: 2}
	}

	number, err := strconv.ParseInt(elements[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse error %w", err)
	}

	options := []HistOption{
		SetID(number),
	}

	commandElements := strings.SplitN(elements[1], "#", 2)
	if len(commandElements) > 1 {
		annotation := strings.Trim(commandElements[1], " ")
		options = append(options, SetAnnotation(annotation))
	}

	history, err := NewHistory(strings.Trim(commandElements[0], " "), options...)
	if err != nil {
		return nil, err
	}

	return history, nil
}

// ConvertOpt creates a HistoryOpt from the history input string
func ConvertOpt(input string) (HistOption, error) {
	trimmed := strings.Trim(input, " ")
	elements := strings.SplitN(trimmed, " ", 2)
	if len(elements) != 2 {
		return nil, ErrIncorrectCount{ArgumentCount: len(elements), ExpectedCount: 2}
	}
	number, err := strconv.ParseInt(elements[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse error %w", err)
	}

	return func(history *History) error {
		history.ID = number
		history.Data = strings.Trim(elements[1], " ")
		return nil
	}, nil
}

func (h History) String() string {
	return fmt.Sprintf("[%s] %s (%s) /*%s*/", h.Time.Format(time.RFC3339), h.Data, h.DirectoryName, h.Annotation)
}

// Store bolddb structure
type Store struct {
	db            *bolt.DB
	directoryFunc func() (string, error)
	timeFunc      func() time.Time
}

// Close on stores
func (s *Store) Close() {
	s.db.Close()
}

// Dump stores
func (s *Store) Dump() {
	history, err := s.All()
	if err != nil {
		logrus.Errorf("could not dump bucket %v\n", err)
		return
	}
	for _, history := range history {
		fmt.Printf("%#v\n", history)
	}
}

type StoreOption func(s *Store) error

func defaultStoreCwd(s *Store) error {
	s.directoryFunc = os.Getwd
	return nil
}

func defaultStoreTime(s *Store) error {
	s.timeFunc = time.Now
	return nil
}

var defaultStoreOptions = []StoreOption{
	defaultStoreTime,
	defaultStoreCwd,
}

// NewStore to create a storage file
func NewStore(dbFile string, options ...StoreOption) (*Store, error) {
	db, err := bolt.Open(dbFile, 0777, nil)
	if err != nil {
		logrus.Errorf("could not open (%s) [%v]", dbFile, err)
		return nil, err
	}
	store := &Store{
		db: db,
	}
	for _, defaultOpt := range defaultStoreOptions {
		defaultOpt(store)
	}
	for _, option := range options {
		option(store)
	}
	return store, nil
}

// Add to storage
func (s *Store) Add(history *History) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(history.DirectoryName))
		if err != nil {
			return err
		}
		ts := []byte(TimeToString(history.Time))
		err = b.Put(ts, []byte(history.Data))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(history.Annotation) != 0 {
		err = s.db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(fmt.Sprintf("annotations-%s", history.DirectoryName)))
			if err != nil {
				return err
			}
			ts := []byte(TimeToString(history.Time))
			err = b.Put(ts, []byte(history.Annotation))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("could not add annotation for history: %w", err)
		}
	}

	return nil
}

// Get from storage
func (s *Store) Get(bucket, key string) (*History, error) {
	var result []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		result = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		return nil, err
	}
	var annotation []byte
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(fmt.Sprintf("annotations-%s", bucket)))
		if b == nil {
			return nil
		}
		annotation = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &History{
		Data:       string(result),
		Annotation: string(annotation),
	}, nil
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

// All entries dumped, with optional filter
func (s *Store) All(filters ...FilterFunction) ([]History, error) {
	history := make([]History, 0, 1000)
	filter := func(bucketName []byte, key []byte, value []byte) bool {
		result := true
		for _, filterFunction := range filters {
			result = filterFunction(bucketName, key, value) && result
			if !result {
				break
			}
		}
		return result
	}
	err := s.ForEachBucket(func(name []byte, b *bolt.Bucket) error {
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				keep := filter(name, k, v)
				if !keep {
					return nil
				}
				snapshotTime, err := StringToTime(string(k))
				if err != nil {
					return err
				}
				result := History{
					Time:          snapshotTime,
					Data:          string(v),
					DirectoryName: string(name),
				}
				history = append(history, result)
				return nil
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return history, nil
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

// FilterFunction provides a type for callback functional options
type FilterFunction func(bucketName []byte, key []byte, value []byte) bool

func applyFilters(filters ...FilterFunction) FilterFunction {
	return func(bucketName []byte, key []byte, value []byte) bool {
		result := true
		for _, filter := range filters {
			result = result && filter(bucketName, key, value)
		}
		return result
	}
}

// Last n entries
func (s *Store) Last(directory string, numEntries int, filters ...FilterFunction) ([]History, error) {
	historyList := make([]History, 0, numEntries)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(directory))
		if b == nil {
			return fmt.Errorf("no such bucket as %s", directory)
		}
		c := b.Cursor()
		i := numEntries
		filter := applyFilters(filters...)
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if i <= 0 {
				break
			}
			keep := filter([]byte(directory), k, v)
			if keep {
				timeValue, err := time.Parse(time.RFC3339, string(k))
				if err != nil {
					return err
				}
				annotationBucket := tx.Bucket([]byte(fmt.Sprintf("annotations-%s", directory)))
				var annotation string
				if annotationBucket != nil {
					annotation = string(annotationBucket.Get(k))
				}
				historyValue := History{
					Data:       string(v),
					Time:       timeValue,
					Annotation: annotation,
				}
				historyList = append(historyList, historyValue)
			}
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
