// Created by Pavel Konovalov pkonovalov@orxagrid.com
//
// Real time (key-value) in memory database
//

package rtdb

import (
	"math"
	"strings"
	"sync"
	"time"
)

const (
	PriorityValueQuality          = 1
	PriorityValueQualityTimestamp = 2
)

type IsoDate struct {
	time.Time
}

// Point value structure
type Point struct {
	Timestamp     IsoDate
	TimestampRecv IsoDate
	Value         float32
	Quality       uint32
	HasFreshData  bool
}

// String ToString functionality
func (c IsoDate) String() string {
	return c.Time.Format("2006-01-02 15:04:05.999")
}

func (c *IsoDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`) // remove quotes
	c.Time, err = time.Parse("2006-01-02T15:04:05.999-0700", s)
	if err != nil {
		c.Time = time.Now()
	}
	return
}

func (c *IsoDate) MarshalJSON() ([]byte, error) {
	str := c.Time.Format("2006-01-02T15:04:05.999-0700")
	return []byte("\"" + str + "\""), nil
}

// Rtdb map with RtdbPoint
type Rtdb struct {
	sync.RWMutex
	db map[uint64]Point
}

// NewRtdb constructor
func NewRtdb() *Rtdb {
	return &Rtdb{
		db: make(map[uint64]Point),
	}
}

// IsPointChanged checking if a point has been changed. It depends on priority
func (c *Rtdb) IsPointChanged(key uint64, priority int, point Point, aperture float64) bool {
	c.RLock()
	if pointInDb, exist := c.db[key]; exist {
		c.RUnlock()

		switch priority {
		case PriorityValueQuality:
			if point.Value != pointInDb.Value || point.Quality != pointInDb.Quality {

				c.Lock()
				c.db[key] = point
				c.Unlock()

				return true
			}
		case PriorityValueQualityTimestamp:
			if point.Value != pointInDb.Value || point.Quality != pointInDb.Quality || point.Timestamp != pointInDb.Timestamp {

				c.Lock()
				c.db[key] = point
				c.Unlock()

				return true
			}
		default:
			if math.Abs(float64(point.Value-pointInDb.Value)) > aperture {

				c.Lock()
				c.db[key] = point
				c.Unlock()

				return true
			}
		}
	} else {
		c.RUnlock()

		c.Lock()
		c.db[key] = point
		c.Unlock()

		return true
	}
	return false
}

// Put save point to rtdb
func (c *Rtdb) Put(key uint64, point Point) {
	c.Lock()
	c.db[key] = point
	c.Unlock()
}

// Get Point by key
func (c *Rtdb) Get(key uint64) (Point, bool) {
	c.RLock()
	point, exist := c.db[key]
	c.RUnlock()
	return point, exist
}

// GetFresh Point by key
func (c *Rtdb) GetFresh(key uint64) (Point, bool) {

	c.RLock()
	point, exist := c.db[key]
	c.RUnlock()

	if exist && point.HasFreshData {
		point.HasFreshData = false

		c.Lock()
		c.db[key] = point
		c.Unlock()

		return point, true
	}

	return point, false
}

// GetCopy of Rtdb
func (c *Rtdb) GetCopy() *map[uint64]Point {
	db := make(map[uint64]Point)

	for k, v := range c.db {
		c.RLock()
		db[k] = v
		c.RUnlock()
	}
	return &db
}
