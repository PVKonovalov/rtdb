// Created by Pavel Konovalov pkonovalov@orxagrid.com
//
// Real time (key-value) in memory database
//

package rtdb

import (
	"strings"
	"sync"
	"time"
)

type IsoDate struct {
	time.Time
}

// Point value structure
type Point struct {
	Timestamp IsoDate
	Value     float32
	Quality   uint32
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
func (c *Rtdb) IsPointChanged(key uint64, priority int, point Point) bool {
	c.RLock()
	if pointInDb, exist := c.db[key]; exist {
		c.RUnlock()
		switch priority {
		case 1:
			if point.Value != pointInDb.Value || point.Quality != pointInDb.Quality {
				c.Lock()
				c.db[key] = point
				c.Unlock()
				return true
			}
		case 2:
			if point.Value != pointInDb.Value || point.Quality != pointInDb.Quality || point.Timestamp != pointInDb.Timestamp {
				c.Lock()
				c.db[key] = point
				c.Unlock()
				return true
			}
		default:
			if point.Value != pointInDb.Value {
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
