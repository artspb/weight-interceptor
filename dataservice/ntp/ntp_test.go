package ntp

import (
	"reflect"
	"testing"
	"time"
)

func TestFromHex(t *testing.T) {
	type args struct {
		hex string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{"", args{"1324fdaf"}, parse("2020-03-06T11:49:03+01:00"), false},
		{"", args{"1324fec8"}, parse("2020-03-06T11:53:44+01:00"), false},
		{"", args{"1324fed9"}, parse("2020-03-06T11:54:01+01:00"), false},
		{"", args{"1324ff01"}, parse("2020-03-06T11:54:41+01:00"), false},
		{"", args{"1324ff12"}, parse("2020-03-06T11:54:58+01:00"), false},
		{"", args{"13250019"}, parse("2020-03-06T11:59:21+01:00"), false},
		{"", args{"13250261"}, parse("2020-03-06T12:09:05+01:00"), false},
		{"", args{"1325312d"}, parse("2020-03-06T15:28:45+01:00"), false},
		{"", args{"132557d1"}, parse("2020-03-06T18:13:37+01:00"), false},
		{"", args{"132575e6"}, parse("2020-03-06T20:21:58+01:00"), false},
		{"", args{"132575e2"}, parse("2020-03-06T20:21:54+01:00"), false},
		{"", args{"13261e05"}, parse("2020-03-07T08:19:17+01:00"), false},
		{"", args{"13262868"}, parse("2020-03-07T09:03:36+01:00"), false},
		{"", args{"13262d3c"}, parse("2020-03-07T09:24:12+01:00"), false},
		{"", args{"13264a30"}, parse("2020-03-07T11:27:44+01:00"), false},
		{"", args{"13276584"}, parse("2020-03-08T07:36:36+01:00"), false},
		{"", args{"13278545"}, parse("2020-03-08T09:52:05+01:00"), false},
		{"", args{"1327c72e"}, parse("2020-03-08T14:33:18+01:00"), false},
		{"", args{"1330d384"}, parse("2020-03-15T11:16:20+01:00"), false},
		{"", args{"1330d38d"}, parse("2020-03-15T11:16:29+01:00"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromHex(tt.args.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromHex() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToHex(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{parse("2020-03-06T11:49:03+01:00")}, "1324fdaf"},
		{"", args{parse("2020-03-06T11:53:44+01:00")}, "1324fec8"},
		{"", args{parse("2020-03-06T11:54:01+01:00")}, "1324fed9"},
		{"", args{parse("2020-03-06T11:54:41+01:00")}, "1324ff01"},
		{"", args{parse("2020-03-06T11:54:58+01:00")}, "1324ff12"},
		{"", args{parse("2020-03-06T11:59:21+01:00")}, "13250019"},
		{"", args{parse("2020-03-06T12:09:05+01:00")}, "13250261"},
		{"", args{parse("2020-03-06T15:28:45+01:00")}, "1325312d"},
		{"", args{parse("2020-03-06T18:13:37+01:00")}, "132557d1"},
		{"", args{parse("2020-03-06T20:21:58+01:00")}, "132575e6"},
		{"", args{parse("2020-03-06T20:21:54+01:00")}, "132575e2"},
		{"", args{parse("2020-03-07T08:19:17+01:00")}, "13261e05"},
		{"", args{parse("2020-03-07T09:03:36+01:00")}, "13262868"},
		{"", args{parse("2020-03-07T09:24:12+01:00")}, "13262d3c"},
		{"", args{parse("2020-03-07T11:27:44+01:00")}, "13264a30"},
		{"", args{parse("2020-03-08T07:36:36+01:00")}, "13276584"},
		{"", args{parse("2020-03-08T09:52:05+01:00")}, "13278545"},
		{"", args{parse("2020-03-08T14:33:18+01:00")}, "1327c72e"},
		{"", args{parse("2020-03-15T11:16:20+01:00")}, "1330d384"},
		{"", args{parse("2020-03-15T11:16:29+01:00")}, "1330d38d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToHex(tt.args.t); got != tt.want {
				t.Errorf("ToHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func parse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func Test_calcBaseline(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{
		{name: "", want: baseline},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcBaseline(); got != tt.want {
				t.Errorf("calcBaseline() = %v, want %v", got, tt.want)
			}
		})
	}
}
