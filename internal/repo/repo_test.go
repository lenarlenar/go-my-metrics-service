package repo

import "testing"

type Pair struct {
	a string
	b float64
}

func TestStorage(t *testing.T) {

	memStorage := NewStorage()

	tests := []struct {
		name   string
		values []Pair
		want   float64
	}{
		{
			name:   "simple test #1",
			values: []Pair{{"name", 5.0}, {"name", 6.7}},
			want:   6.7,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, p := range test.values {
				memStorage.SetGauge(p.a, p.b)
			}

			metric := memStorage.GetMetrics()[test.values[0].a]
			if *metric.Value != test.want {
				t.Errorf("test.values[0].a = %f, want %f", *metric.Value, test.want)
			}
		})
	}
}
