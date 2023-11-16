package bwt

import (
	"math/rand"
	"strings"
	"testing"
)

// type QueryTest struct {
// 	seq      string
// 	expected bool
// }
//
// func TestQueryBWT(t *testing.T) {
// 	bwt := New("BANANA")
//
// 	testTable := []QueryTest{
// 		{"NANA", true},
// 		{"ANA", true},
// 		{"NA", true},
// 		{"B", true},
// 		{"N", true},
// 		{"BA", true},
// 		{"ANANA", true},
// 		{"QWERTY", false},
// 		{"ANANANA", false},
// 		{"ABCD", false},
// 		{"ABA", false},
// 	}
//
// 	for _, v := range testTable {
// 		res := bwt.QueryExistence(v.seq)
// 		if res != v.expected {
// 			t.Fatalf("Test=%s ExpectedQueryExistence=%v Received=%v", v.seq, v.expected, res)
// 		}
// 	}
// }
//
// func BenchmarkBWTBuildPower12(b *testing.B) {
// 	base := "!BANANA!"
// 	BaseBenchmarkBWTBuild(base, 12, b)
// }
//
// //go:noinline
// func BaseBenchmarkBWTBuild(base string, power int, b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		buildBWTForBench(base, power)
// 	}
// }
//
// func buildBWTForBench(base string, power int) BWT {
// 	test := base
// 	for i := 0; i < power; i++ {
// 		test += test
// 	}
//
// 	return New(test)
// }
//
// func BenchmarkBWTQueryPower12(b *testing.B) {
// 	base := "!BANANA!"
// 	bwt := buildBWTForBench(base, 12)
// 	BaseBenchmarkBWTQuery(bwt, "ANANABANANA", b)
// }
//
// //go:noinline
// func BaseBenchmarkBWTQuery(bwt BWT, seq string, b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		bwt.QueryExistence(seq)
// 	}
// }

func genRandStrs(seed int64, alpha string) []string {
	r := rand.New(rand.NewSource(seed))
	minLen := 50
	maxLen := 500
	numStrs := 10000
	strs := make([]string, numStrs)
	for i := 0; i < numStrs; i++ {
		ln := minLen + r.Intn(maxLen-minLen)
		str := strings.Builder{}
		for j := 0; j < ln; j++ {
			str.WriteByte(alpha[rand.Intn(len(alpha))])
		}
		strs[i] = str.String()
	}

	return strs
}

// allowedChars holds a map of allowed characters for DNA/RNA, used by Hash2Fragment:
var allowedChars = map[rune]bool{'A': true, 'T': true, 'U': true, 'G': true, 'C': true, 'Y': true, 'R': true, 'S': true, 'W': true, 'K': true, 'M': true, 'B': true, 'D': true, 'H': true, 'V': true, 'N': true, 'Z': true}

func BenchmarkMapSmallAlphaWithNoMistakes(b *testing.B) {
	alpha := "ATGC"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkMap(b, strs)
}

func BenchmarkMapSmallAlphaWithSomeMistakes(b *testing.B) {
	alpha := "ATGC+["
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkMap(b, strs)
}

func BenchmarkMapSmallAlphaWithManyMistakes(b *testing.B) {
	alpha := "ATGC+[(=}*]})"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkMap(b, strs)
}

func BenchmarkMapCompleteAlphaWithNoMistakes(b *testing.B) {
	alpha := "ATUGCYRSWKMBDHVNZ"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkMap(b, strs)
}

func BenchmarkMapCompleteAlphaWithSomeMistakes(b *testing.B) {
	alpha := "ATUGCYRSWKMBDHVNZ+[()}*"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkMap(b, strs)
}

func BenchmarkMapCompleteAlphaWithManyMistakes(b *testing.B) {
	alpha := "ATUGCYRSWKMBDHVNZ+[()}*@!$^%1234567890abcdefg"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkMap(b, strs)
}

func BenchmarkMapCompleteAlphaOopsAllMistakes(b *testing.B) {
	alpha := "abcdefghijklmnopqrstuvwxyz1234567890+[{(&=)}]*$!#%^@"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkMap(b, strs)
}

//go:noinline
func BaseBenchmarkMap(b *testing.B, testStrs []string) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(testStrs); i++ {
			for _, c := range testStrs[i] {
				_, _ = allowedChars[c]
			}
		}
	}
}

func BenchmarkContainsSmallAlphaWithNoMistakes(b *testing.B) {
	alpha := "ATGC"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkContains(b, strs)
}

func BenchmarkContainsSmallAlphaWithSomeMistakes(b *testing.B) {
	alpha := "ATGC+["
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkContains(b, strs)
}

func BenchmarkContainsSmallAlphaWithManyMistakes(b *testing.B) {
	alpha := "ATGC+[(=}*]})"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkContains(b, strs)
}

func BenchmarkContainsCompleteAlphaWithNoMistakes(b *testing.B) {
	alpha := "ATUGCYRSWKMBDHVNZ"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkContains(b, strs)
}

func BenchmarkContainsCompleteAlphaWithSomeMistakes(b *testing.B) {
	alpha := "ATUGCYRSWKMBDHVNZ+[()}*"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkContains(b, strs)
}

func BenchmarkContainsCompleteAlphaWithManyMistakes(b *testing.B) {
	alpha := "ATUGCYRSWKMBDHVNZ+[()}*@!$^%1234567890abcdefg"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkContains(b, strs)
}

func BenchmarkContainsCompleteAlphaOopsAllMistakes(b *testing.B) {
	alpha := "abcdefghijklmnopqrstuvwxyz1234567890+[{(&=)}]*$!#%^@"
	strs := genRandStrs(12345, alpha)
	BaseBenchmarkContains(b, strs)
}

//go:noinline
func BaseBenchmarkContains(b *testing.B, testStrs []string) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < len(testStrs); i++ {
			for j := 0; j < len(testStrs[i]); j++ {
				strings.Contains("ATUGCYRSWKMBDHVNZ", string(testStrs[i][j]))
			}
		}
	}
}
