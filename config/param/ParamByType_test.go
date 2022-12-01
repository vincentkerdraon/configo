package param

import (
	"fmt"
	"testing"
	"time"
)

func TestNewPlusSpecificType(t *testing.T) {
	var res string
	p, _ := NewString("name", func(s string) error {
		res = fmt.Sprintf("%T:%v", s, s)
		return nil
	})
	if err := p.Parse("10"); err != nil {
		t.Error(err)
	}
	p, _ = NewBool("name", func(b bool) error {
		res = fmt.Sprintf("%s %T:%v", res, b, b)
		return nil
	})
	if err := p.Parse("true"); err != nil {
		t.Error(err)
	}
	p, _ = NewInt("name", func(i int) error {
		res = fmt.Sprintf("%s %T:%v", res, i, i)
		return nil
	})
	if err := p.Parse("10"); err != nil {
		t.Error(err)
	}
	p, _ = NewInt64("name", func(i int64) error {
		res = fmt.Sprintf("%s %T:%v", res, i, i)
		return nil
	})
	if err := p.Parse("10"); err != nil {
		t.Error(err)
	}
	p, _ = NewUint("name", func(i uint) error {
		res = fmt.Sprintf("%s %T:%v", res, i, i)
		return nil
	})
	if err := p.Parse("10"); err != nil {
		t.Error(err)
	}
	p, _ = NewUint64("name", func(i uint64) error {
		res = fmt.Sprintf("%s %T:%v", res, i, i)
		return nil
	})
	if err := p.Parse("10"); err != nil {
		t.Error(err)
	}
	p, _ = NewFloat64("name", func(f float64) error {
		res = fmt.Sprintf("%s %T:%v", res, f, f)
		return nil
	})
	if err := p.Parse("10.1"); err != nil {
		t.Error(err)
	}
	p, _ = NewDuration("name", func(d time.Duration) error {
		res = fmt.Sprintf("%s %T:%v", res, d, d)
		return nil
	})
	if err := p.Parse("10s"); err != nil {
		t.Error(err)
	}
	expected := `string:10 bool:true int:10 int64:10 uint:10 uint64:10 float64:10.1 time.Duration:10s`
	if res != expected {
		t.Errorf("\ngot =%v\nwant=%v", res, expected)
	}
}
