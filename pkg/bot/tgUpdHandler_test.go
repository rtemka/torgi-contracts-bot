package bot

import (
	botDB "tbot/pkg/db"
	"testing"
)

type dbQueryManagerMock struct {
}

func (qm dbQueryManagerMock) Query(_ int, _ ...botDB.QueryOpt) ([]botDB.PurchaseRecord, error) {
	return nil, nil
}

func (qm dbQueryManagerMock) QueryRow(_ int64) (botDB.PurchaseRecord, error) {
	return botDB.PurchaseRecord{}, nil
}

func TestTgUpdHandlerParseFlags(t *testing.T) {

	t.Run("parse_flags_empty", func(t *testing.T) {
		got, err := parseFlags(nil)
		if err != nil {
			t.Fatalf("due parsing flags [%v]", err)
		}

		if *got != (flags{set: got.set}) {
			t.Fatalf("expected empty flag set, got [%v]", got)
		}
	})

	t.Run("parse_flags_short", func(t *testing.T) {
		args := []string{"-a", "-g", "-m", "-d", "1", "-t", "-f", "-p"}
		got, err := parseFlags(args)
		if err != nil {
			t.Fatalf("due parsing flags [%v]", err)
		}

		expect := flags{set: got.set, tf: true, pf: true, ff: true, af: true, gf: true, mf: true, df: 1}

		if *got != expect {
			t.Fatalf("from %v flags, expected %+v flag set, got %+v", args, expect, got)
		}

		args = []string{"--a", "--g", "--m", "--d", "1", "-t", "-f", "-p"}
		got, err = parseFlags(args)
		if err != nil {
			t.Fatalf("due parsing flags [%v]", err)
		}

		expect.set = got.set
		if *got != expect {
			t.Fatalf("from %v flags, expected %+v flag set, got %+v", args, expect, got)
		}
	})

	t.Run("parse_flags_long", func(t *testing.T) {
		args := []string{"-auction", "-go", "-money", "-days", "1", "-t", "-f", "-p"}
		got, err := parseFlags(args)
		if err != nil {
			t.Fatalf("due parsing flags [%v]", err)
		}

		expect := flags{set: got.set, tf: true, pf: true, ff: true, af: true, gf: true, mf: true, df: 1}

		if *got != expect {
			t.Fatalf("from %v flags, expected %+v flag set, got %+v", args, expect, got)
		}

		args = []string{"--auction", "--go", "--money", "--days", "1", "-t", "-f", "-p"}
		got, err = parseFlags(args)
		if err != nil {
			t.Fatalf("due parsing flags [%v]", err)
		}

		expect.set = got.set
		if *got != expect {
			t.Fatalf("from %v flags, expected %+v flag set, got %+v", args, expect, got)
		}
	})
}
