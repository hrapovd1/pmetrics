package storage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage_Rewrite(t *testing.T) {
	type args struct {
		key   string
		value gauge
	}
	tests := []args{
		{
			key:   "1",
			value: -0.1,
		},
		{
			key:   "1",
			value: 1,
		},
	}
	ms := NewMemStorage()
	t.Run("Rewrite values.", func(t *testing.T) {
		for _, tt := range tests {
			ms.Rewrite(tt.key, tt.value)
			assert.Equal(t, tt.value, ms.GaugeBuff[tt.key])
		}
	})
	t.Run("Count values.", func(t *testing.T) {
		assert.Equal(t, 1, len(ms.GaugeBuff))
	})
}

func TestMemStorage_Append(t *testing.T) {
	tests := []counter{1, 2, 3}
	ms := NewMemStorage()
	t.Run("Append values", func(t *testing.T) {
		for _, val := range tests {
			ms.Append(val)
			last := len(ms.CounterBuff) - 1
			assert.Equal(t, val, ms.CounterBuff[last])
		}
	})
	t.Run("Count values", func(t *testing.T) {
		assert.Equal(t, 3, len(ms.CounterBuff))
	})
}

func TestNewMemStorage(t *testing.T) {
	want := &MemStorage{
		GaugeBuff:   make(map[string]gauge),
		CounterBuff: make([]counter, 0),
	}
	assert.True(t, cmp.Equal(NewMemStorage(), want))
}

func TestStrToGauge(t *testing.T) {
	tests := []struct {
		name     string
		positive bool
		value    string
		want     gauge
	}{
		{
			name:     "Float value",
			positive: true,
			value:    "1.0",
			want:     gauge(1.0),
		},
		{
			name:     "Int value",
			positive: true,
			value:    "1",
			want:     gauge(1.0),
		},
		{
			name:     "Wrong value",
			positive: false,
			value:    "a",
			want:     gauge(1.0),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.positive {
				value, err := StrToGauge(test.value)
				assert.Nil(t, err)
				assert.Equal(t, test.want, value)
			} else {
				_, err := StrToGauge(test.value)
				assert.NotNil(t, err)
			}
		})
	}
}

func TestStrToCounter(t *testing.T) {
	tests := []struct {
		name     string
		positive bool
		value    string
		want     counter
	}{
		{
			name:     "Float value",
			positive: false,
			value:    "1.0",
			want:     counter(1),
		},
		{
			name:     "Int value",
			positive: true,
			value:    "-0",
			want:     counter(0),
		},
		{
			name:     "Wrong value",
			positive: false,
			value:    "a",
			want:     counter(0),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.positive {
				value, err := StrToCounter(test.value)
				assert.Nil(t, err)
				assert.Equal(t, test.want, value)
			} else {
				_, err := StrToCounter(test.value)
				assert.NotNil(t, err)
			}
		})
	}
}
