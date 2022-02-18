package gsa

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"birc.au.dk/gsa/test"
)

func TestOpsToCigar(t *testing.T) {
	type args struct {
		ops EditOps
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Single M",
			args{ops: []ApproxEdit{M}},
			"1M",
		},
		{
			"Single D",
			args{ops: []ApproxEdit{D}},
			"1D",
		},
		{
			"Single I",
			args{ops: []ApproxEdit{I}},
			"1I",
		},
		{
			"IIMMMDDI",
			args{ops: []ApproxEdit{
				I, I, M, M, M, D, D, I}},
			"2I3M2D1I",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OpsToCigar(tt.args.ops); got != tt.want {
				t.Errorf("OpsToCigar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCigarToOps(t *testing.T) {
	tests := []struct {
		cigar   string
		want    EditOps
		wantErr error
	}{
		{
			"1M",
			EditOps{M},
			nil,
		},
		{
			"10M",
			EditOps{M, M, M, M, M, M, M, M, M, M},
			nil,
		},
		{
			"1I",
			EditOps{I},
			nil,
		},
		{
			"1D",
			EditOps{D},
			nil,
		},
		{
			"1D2M3I",
			EditOps{D, M, M, I, I, I},
			nil,
		},
		{
			"invalid",
			EditOps{},
			NewInvalidCigar("invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.cigar, func(t *testing.T) {
			got, gotErr := CigarToOps(tt.cigar)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("Unexpected error, %q", gotErr)
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CigarToOps() = %v, want %v", got, tt.want)
			}

			if gotErr != nil && gotErr.Error() != fmt.Sprintf("invalid cigar: %s", tt.cigar) {
				t.Errorf("Unexpected error message: %s", gotErr)
			}
		})
	}
}

func TestExtractAlignment(t *testing.T) {
	type args struct {
		x     string
		p     string
		pos   int32
		cigar string
	}

	tests := []struct {
		name     string
		args     args
		wantSubx string
		wantSubp string
		wantErr  error
	}{
		{
			"Just matches",
			args{"acgtacgt", "gtac", 2, "4M"},
			"gtac", "gtac",
			nil,
		},
		{
			"Deletion",
			args{"acgtacgt", "gtc", 2, "2M1D1M"},
			"gtac", "gt-c",
			nil,
		},
		{
			"Insertion",
			args{"acgtacgt", "gtaac", 2, "2M1I2M"},
			"gt-ac", "gtaac",
			nil,
		},
		{
			"Invalid",
			args{"acgtacgt", "gtaac", 2, "invalid"},
			"", "",
			NewInvalidCigar("invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSubx, gotSubp, gotErr := ExtractAlignment(tt.args.x, tt.args.p, tt.args.pos, tt.args.cigar)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Fatalf("ExtractAlignment() gotErr = %v, want %v", gotErr, tt.wantErr)
			}
			if gotSubx != tt.wantSubx {
				t.Errorf("ExtractAlignment() gotSubx = %v, want %v", gotSubx, tt.wantSubx)
			}
			if gotSubp != tt.wantSubp {
				t.Errorf("ExtractAlignment() gotSubp = %v, want %v", gotSubp, tt.wantSubp)
			}
		})
	}
}

func TestCountEdits(t *testing.T) {
	type args struct {
		x     string
		p     string
		pos   int32
		cigar string
	}

	tests := []struct {
		name    string
		args    args
		want    int
		wantErr error
	}{
		{
			"Just matches",
			args{"acgtacgt", "gtac", 2, "4M"},
			0, // "gtac" vs "gtac",
			nil,
		},
		{
			"Deletion",
			args{"acgtacgt", "gtc", 2, "2M1D1M"},
			1, // "gtac", "gt-c",
			nil,
		},
		{
			"Insertion",
			args{"acgtacgt", "gtaac", 2, "2M1I2M"},
			1, // "gt-ac", "gtaac",
			nil,
		},
		{
			"Invalid",
			args{"acgtacgt", "gtaac", 2, "invalid"},
			0, // error...
			NewInvalidCigar("invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := CountEdits(tt.args.x, tt.args.p, tt.args.pos, tt.args.cigar)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Fatalf("Unexpected error %v", gotErr)
			}
			if got != tt.want {
				t.Errorf("CountEdits() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type approxAlgo = func(string) func(string, int, func(int32, string))

var approxAlgorithms = map[string]approxAlgo{
	"BWA": FMIndexApproxPreprocess,
}

func runRandomApproxOccurencesTests(algo approxAlgo) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		rng := test.NewRandomSeed(t)
		test.GenerateTestStringsAndPatterns(10, 20, rng,
			func(x, p string) {
				search := algo(x)
				for edits := 1; edits < 3; edits++ {
					search(p, edits, func(pos int32, cigar string) {
						count, _ := CountEdits(x, p, pos, cigar)
						if count > edits {
							fmt.Println(pos, cigar)
							ax, ap, _ := ExtractAlignment(x, p, pos, cigar)
							fmt.Printf("%s\n%s\n\n", ax, ap)

							t.Errorf("Match at pos %d needs too many edits, %d vs %d",
								pos, count, edits)
						}
					})
				}
			})
	}
}

func TestRandomApproxOccurences(t *testing.T) {
	t.Helper()

	for name, algo := range approxAlgorithms {
		t.Run(name, runRandomApproxOccurencesTests(algo))
	}
}
