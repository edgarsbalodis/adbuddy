package scraper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDate(t *testing.T) {
	rigaLocation, _ := time.LoadLocation("Europe/Riga")
	wantTime, err := time.ParseInLocation("02.01.2006 15:04:00 Europe/Riga", "13.10.2023 11:15:00 Europe/Riga", rigaLocation)
	if err != nil {
		t.Fatalf("Error parsing time: %v", err)
	}
	type args struct { // function args definition
		a string // text from ss.com date field
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{ // test cases
		{ // test case
			name: "should return formatted date and time",
			// args: args{"Datums: 13.10.2023 11:15", "Datums: 14.10.2023 23:24", "Datums: 14.10.2023 10:33"},
			args: args{"Datums: 13.10.2023 11:15"},
			want: wantTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDate(tt.args.a) // function call
			assert.Equal(t, tt.want, got)
		})
	}
}
