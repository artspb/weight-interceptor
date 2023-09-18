package parser

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
	"weight-interceptor-http/dataservice/hash"
	"weight-interceptor-http/dataservice/ntp"
)

func Test_parseData(t *testing.T) {
	type args struct {
		request string
	}
	tests := []struct {
		name    string
		args    args
		want    Weight
		wantErr bool
	}{
		{name: "empty request", args: args{""}, want: Weight{}, wantErr: true},
		{name: "too short request", args: args{strings.Repeat("0", 67)}, want: Weight{}, wantErr: true},
		{name: "too long request", args: args{strings.Repeat("0", 69)}, want: Weight{}, wantErr: true},
		{name: "checksum doesn't match", args: args{code + strings.Repeat("0", 66)}, want: Weight{}, wantErr: true},
		{name: "checksum matches", args: args{withChecksum(code + strings.Repeat("0", 58))}, want: Weight{strings.Repeat("0", 24), toTime("2010-01-01 00:00:00 +0100 CET"), 0}, wantErr: false},
		{name: "uid", args: args{withChecksum(code + strings.Repeat("1", 24) + strings.Repeat("0", 34))}, want: Weight{strings.Repeat("1", 24), toTime("2010-01-01 00:00:00 +0100 CET"), 0}, wantErr: false},
		{name: "", args: args{message("2010-01-01 00:00:00 +0100 CET", 0)}, want: Weight{strings.Repeat("0", 24), toTime("2010-01-01 00:00:00 +0100 CET"), 0}, wantErr: false},
		{name: "", args: args{message("2015-01-01 00:00:00 +0100 CET", 100)}, want: Weight{strings.Repeat("0", 24), toTime("2015-01-01 00:00:00 +0100 CET"), 100}, wantErr: false},
		{name: "", args: args{withChecksum("2400000000000000000000000000001330a2fe1cfc000000000000000000")}, want: Weight{strings.Repeat("0", 24), toTime("2020-03-15 07:49:18 +0100 CET"), 7420}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseData(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func message(t string, w int) string {
	return withChecksum(code + strings.Repeat("0", 28) + ntp.ToHex(toTime(t)) + fmt.Sprintf("%04x", w) + strings.Repeat("0", 18))
}

func withChecksum(message string) string {
	checksum, err := hash.Checksum(message)
	if err != nil {
		panic(err)
	}
	return message + checksum
}

func toTime(t string) time.Time {
	result, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", t)
	if err != nil {
		panic(err)
	}
	return result
}
