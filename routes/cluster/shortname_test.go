package cluster

import (
	"testing"
)

func TestGenerateShortName(t *testing.T) {
	type args struct {
		name              string
		maxSize           int
		versionsPerOption int
		filter            []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				name:              "gke_elated-pottery-310110_us-central1-c_dwertent",
				maxSize:           5,
				versionsPerOption: 3,
				filter:            nil,
			},
			want: "D",
		},
		{
			name: "test2",
			args: args{
				name:              "gke_elated-pottery-310110_us-central1-c_dwertent",
				maxSize:           5,
				versionsPerOption: 3,
				filter:            []string{"D"},
			},
			want: "DW",
		},
		{
			name: "test3",
			args: args{
				name:              "gke_elated-pottery-310110_us-central1-c_dwertent",
				maxSize:           5,
				versionsPerOption: 3,
				filter:            []string{"D", "DW", "DWER", "DWERT"},
			},
			want: "DWE",
		},
		{
			name: "test4",
			args: args{
				name:              "gke_elated-pottery-310110_us-central1-c_dwertent",
				maxSize:           5,
				versionsPerOption: 3,
				filter:            []string{"D", "DW", "DWE", "DWER", "DWERT"},
			},
			want: "D1",
		},
		{
			name: "test5",
			args: args{
				name:              "gke_elated-pottery-310110_us-central1-c_dwertent",
				maxSize:           5,
				versionsPerOption: 3,
				filter:            []string{"D", "DW", "DWE", "DWER", "DWERT", "D1", "D2", "DW1"},
			},
			want: "DW2",
		},
		{
			name: "test6",
			args: args{
				name:              "gke_elated-pottery",
				maxSize:           5,
				versionsPerOption: 0,
				filter:            []string{"P", "PO"},
			},
			want: "POT",
		},
		{
			name: "test7",
			args: args{
				name:              "77-gke_elated-pottery-12",
				maxSize:           5,
				versionsPerOption: 0,
				filter:            []string{"P", "PO"},
			},
			want: "POT",
		},
		{
			name: "test8",
			args: args{
				name:              "gke_elated-pot",
				maxSize:           5,
				versionsPerOption: 3,
				filter:            []string{"P", "PO", "POT", "G", "GK", "GKE", "GKEE", "GKEEP", "P1"},
			},
			want: "P2",
		},
		{
			name: "test9",
			args: args{
				name:              "m",
				maxSize:           5,
				versionsPerOption: 5,
				filter:            []string{"M", "M1", "M2", "M3"},
			},
			want: "M4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := longName2short(tt.args.name, tt.args.maxSize, tt.args.versionsPerOption, tt.args.filter); got != tt.want {
				//if got := GenerateAcronym(tt.args.name, tt.args.maxSize, tt.args.versionsPerOption, tt.args.filter); got != tt.want {
				t.Errorf("%s longName2short() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}

}
