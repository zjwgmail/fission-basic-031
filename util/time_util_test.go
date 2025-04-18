package util

import (
	"context"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	time, err := IsNotDisturbTime(context.Background(), false, "38057481920")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(time)
}

func TestGetSendClusteringTime(t *testing.T) {
	type args struct {
		ctx       context.Context
		waId      string
		afterHour int
		nowUnix   int64
	}

	testTime := time.Date(25, 2, 27, 20, 59, 59, 59, time.Local)
	nowUnix := testTime.Unix()
	want := testTime.Add(3 * time.Hour).Unix()

	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
		// "60":  "Asia/Kuala_Lumpur",
		// "62":  "Asia/Jakarta",
		// "63":  "Asia/Manila",
		// "65":  "Asia/Singapore",
		// "66":  "Asia/Bangkok",
		// "84":  "Asia/Ho_Chi_Minh",
		// "7":   "Europe/Moscow",
		// "90":  "Europe/Istanbul",
		// "966": "Asia/Riyadh",
		// "380": "Europe/Kiev",
		// "375": "Europe/Minsk",
		// "998": "Asia/Tashkent",
		// "996": "Asia/Bishkek",
		// "994": "Asia/Baku",
		// "373": "Europe/Chisinau",
		// "992": "Asia/Dushanbe",
		// "374": "Asia/Yerevan",
		// "971": "Asia/Dubai",
		// "973": "Asia/Bahrain",
		// "974": "Asia/Qatar",
		// "965": "Asia/Kuwait",
		// "968": "Asia/Muscat",
		// "20":  "Africa/Cairo",
		// "216": "Africa/Tunis",
		// "213": "Africa/Algiers",
		// "92":  "Asia/Karachi",
		// "880": "Asia/Dhaka",
		// "852": "Asia/Hong_Kong",
		{
			name: "case1",
			args: args{
				ctx:       context.Background(),
				waId:      "60057481920",
				afterHour: 3,
				nowUnix:   nowUnix,
			},
			want: want,
		},
		{
			name: "case2",
			args: args{
				ctx:       context.Background(),
				waId:      "62057481920",
				afterHour: 3,
				nowUnix:   nowUnix,
			},
			want: want,
		},
		{
			name: "case3",
			args: args{
				ctx:       context.Background(),
				waId:      "63057481920",
				afterHour: 3,
				nowUnix:   nowUnix,
			},
			want: want,
		},
		{
			name: "case4",
			args: args{
				ctx:       context.Background(),
				waId:      "65057481920",
				afterHour: 3,
				nowUnix:   nowUnix,
			},
			want: want,
		},
		{
			name: "case5",
			args: args{
				ctx:       context.Background(),
				waId:      "66057481920",
				afterHour: 3,
				nowUnix:   1650000000,
			},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSendClusteringTime(tt.args.ctx, tt.args.waId, tt.args.afterHour, tt.args.nowUnix)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSendClusteringTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSendClusteringTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
