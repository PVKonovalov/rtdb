# rtdb
Simple goroutine-safe realtime (in memory) database for SCADA application

Each point stored in RTDB has the following fields:

```
type Point struct {
	Timestamp     IsoDate // Timestamp from remote terminal unit (RTU) 
	TimestampRecv IsoDate // Timestamp from SCADA
	Value         float32 // Value
	Quality       uint32  // Quality descriptor: bitstring 
	HasFreshData  bool    // Flag indicating that the data has changed
}
```
