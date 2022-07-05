package bot

import (
	"reflect"
	"testing"
)

// func Test_parseFlags1(t *testing.T) {

// 	t.Run("empty", func(t *testing.T) {
// 		got, err := parseFlags(nil)
// 		if err != nil {
// 			t.Fatalf("due parsing flags [%v]", err)
// 		}

// 		if *got != (flags{set: got.set}) {
// 			t.Fatalf("expected empty flag set, got [%v]", got)
// 		}
// 	})

// 	t.Run("short", func(t *testing.T) {
// 		args := []string{"-a", "-g", "-m", "-d", "1", "-t", "-f", "-p"}
// 		got, err := parseFlags(args)
// 		if err != nil {
// 			t.Fatalf("due parsing flags [%v]", err)
// 		}

// 		expect := flags{set: got.set, tf: true, pf: true, ff: true, af: true, gf: true, mf: true, df: 1}

// 		if *got != expect {
// 			t.Fatalf("from %v flags, expected %+v flag set, got %+v", args, expect, got)
// 		}

// 		args = []string{"--a", "--g", "--m", "--d", "1", "-t", "-f", "-p"}
// 		got, err = parseFlags(args)
// 		if err != nil {
// 			t.Fatalf("due parsing flags [%v]", err)
// 		}

// 		expect.set = got.set
// 		if *got != expect {
// 			t.Fatalf("from %v flags, expected %+v flag set, got %+v", args, expect, got)
// 		}
// 	})

// 	t.Run("long", func(t *testing.T) {
// 		args := []string{"-auction", "-go", "-money", "-days", "1", "-t", "-f", "-p"}
// 		got, err := parseFlags(args)
// 		if err != nil {
// 			t.Fatalf("due parsing flags [%v]", err)
// 		}

// 		expect := flags{set: got.set, tf: true, pf: true, ff: true, af: true, gf: true, mf: true, df: 1}

// 		if *got != expect {
// 			t.Fatalf("from %v flags, expected %+v flag set, got %+v", args, expect, got)
// 		}

// 		args = []string{"--auction", "--go", "--money", "--days", "1", "-t", "-f", "-p"}
// 		got, err = parseFlags(args)
// 		if err != nil {
// 			t.Fatalf("due parsing flags [%v]", err)
// 		}

// 		expect.set = got.set
// 		if *got != expect {
// 			t.Fatalf("from %v flags, expected %+v flag set, got %+v", args, expect, got)
// 		}
// 	})
// }

func Test_parseFlags(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		args    args
		want    *flags
		wantErr bool
	}{
		{
			name:    "empty",
			want:    &flags{},
			args:    args{[]string{}},
			wantErr: false,
		},
		{
			name:    "short",
			want:    &flags{set: nil, tf: true, pf: true, ff: true, af: true, gf: true, mf: true, df: 1},
			args:    args{[]string{"-a", "-g", "-m", "-d", "1", "-t", "-f", "-p"}},
			wantErr: false,
		},
		{
			name:    "long",
			want:    &flags{set: nil, tf: true, pf: true, ff: true, af: true, gf: true, mf: true, df: 1},
			args:    args{[]string{"-auction", "-go", "-money", "-days", "1", "-t", "-f", "-p"}},
			wantErr: false,
		},
		{
			name:    "long_double_dash",
			want:    &flags{set: nil, tf: true, pf: true, ff: true, af: true, gf: true, mf: true, df: 1},
			args:    args{[]string{"--auction", "--go", "--money", "--days", "1", "-t", "-f", "-p"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlags(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.want.set = got.set
			if !reflect.DeepEqual(*got, *tt.want) {
				t.Errorf("parseFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}
