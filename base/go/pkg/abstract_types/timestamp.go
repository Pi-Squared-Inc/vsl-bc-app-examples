package abstract_types

type Timestamp struct {
	Seconds uint64 `json:"seconds"`
	Nanos   uint32 `json:"nanos"`
}

func (t *Timestamp) Tick() {
	if t.Nanos == 999_999_999 {
		t.Nanos = 0
		t.Seconds += 1
	} else {
		t.Nanos++
	}
}
