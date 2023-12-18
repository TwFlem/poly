package bwt

import (
	"fmt"
	"math"
	"strings"

	"golang.org/x/exp/slices"
)

/*

For the BWT usage, please read the its
method documentation. To understand what it is and how
it works for either curiosity or maintenance, then read below.

# BWT

BWT Stand for (B)urrow (W)heeler (T)ransform. The BWT aids in
text compression and acts as a search index for any arbitrary
sequence of characters. With the BWT and some auxiliary data
structures, we can analyze a sequence in a memory and run time
efficient manner.

## BWT Transform

The first step to build the BWT is to get the BWT itself.

This is done by:
1. Appending a null terminating character to the end of a sequence
2. Rotate the sequence so that the last character is now the first
3. Repeat 2. N times where N is the length of the sequence
4. Lexicographically sort the NxN matrix of rotated sequences where
   the null termination character is always the least-valued
5. Store the first and last column of the matrix. The last column
   is the output of the BWT. The first column is needed to run queries
   on the BWT of the original sequence.

Lets use banana as an example.

banana$      $banana
$banana      a$banan
a$banan      ana$ban
na$bana  =>  anana$b
ana$ban      banana$
nana$ba      na$bana
anana$b      nana$ba

Output:

Last Column (BWT): annb$aa
First Column:     $aaabnn

## LF Mapping Properties

From now on we will refer to the Last Column as L and the First as F

There are a few special properties here to be aware of. First, notice
how the characters of the same rank show up in the same order for each
column:

L: a0 n0 n1 b0 $0 a1 a2

F: $0 a0 a1 a2 b0 n0 n1

That is to say the characters' rank for each column appear in ascending
order. For example: a0 < a1 < a2. This is true for all BWTs

The other important property to observe is that since the BWT is the
result of rotating each string, each character in the L column precedes
the corresponding character in the F column.

To best way to show this is to rebuild the original sequence
using the F and L columns. We do this by rebuilding the original string
in reverse order starting with the nullChar.

Original string: ______$0

F($0) -> L(a0) -> _____a0$0
F(a0) -> L(n0) -> ____n0a0$0
F(n0) -> L(a1) -> ___a1n0a0$0
F(a1) -> L(n1) -> __n1a1n0a0$0
F(n1) -> L(a2) -> _a2n1a1n0a0$0
F(a2) -> L(b0) -> b0a2n1a1n0a0$0
F(b0) -> L($0) -> Complete

If we take the rank subscripts away from: b0a2n1a1n0a0$0
We get... "banana$" !

## LF Mapping Usage

From these properties, the most important concept emerges- the LF Mapping.
The LF mapping is what enables us to query and analyze the BWT to gain
insight about the original sequence.

For example, let's say we wanted to count the number of occurrences of the
pattern "ana" in "banana". We can do this by:

1. Lookup the last char of the sequence, a, in the F column
2. Find that range of a's, [1, 4)
3. Take the next previous character in the pattern, n
4. Find the rank of n before the range from 2. [0, 1) = 0
5. Find the rank of n in the range from 2. [1, 4) = 1
6. Look up the start range of the n's in the F column, 5
7. Add the result from 4 and 5 respectively to form the next
   L search range: [5+0, 5+1) = [5, 6)
8. Take next previous character in the pattern, a
9. Take the rank of "a" before, 0
10. Take the rank of "a" within, 1
11. Lookup the a's in the F column again, but add the results
    from 9 and 10 to the start range to get the next search
    range = [1+0, 1+1) = [1, 2)
12. That is beginning of out pattern, we sub subtract the end and start
    of the search range to get out count, 2-1=1

Another way to look at this is that we are constantly refining our search
range for each character of the pattern we are searching for. Once we
reach the end of the pattern, our final range represents the a's which
start our pattern. If the range < 0, then at some point our search ranged
has collapsed and we can conclude that there is no matching pattern.

## Suffix Array

For other operations such as Locate and Extract, we need another auxiliary
data structure, the suffix array. Since we could be at multiple points
within the original sequence and at any point within that sequence, we need
some kind of point of reference of where we are. We can do this by storing
the position of each original character for each of the corresponding
characters in the F column. With our banana example:

F:   $0 a0 a1 a2 b0 n0 n1
SA: [6  5  3  1  0  4  2]

If we take our count example for the pattern "ana" above, you'll remember
that our final search range was [1, 2). If we look up 1 in the SA, we'll
fund that there is only one offset at position 3 in the original sequence
"banana"

## Notes on Performance

The explanation above leads to a very naive implementation. For example,
having the full SA would take way more memory than the BWT itself. Assuming
int64, that would 8 times the amount of memory of the BWT in its plain text
representation! In the implementation below, we may instead sample the SA
and do additional look ups as needed to find the offsets we need.

Similarly, storing the F and L column as plain text has just doubled the
amount of memory from the original sequence... BWT is used for text
compression, not expansion! That's why in the below implementation, you
will see other data structures to actually compress the amount of memory
needed. You will also notice that we can make huge improvements by
compressing sequences like with the F column.

Instead of:

F: $0 a0 a1 a2 b0 n0 n1

Since F is lexicographically sorted, we can have:

F: {$: [0, 1)}, {a: [1, 4)}, {b: [4, 5)} {n: [5, 7)}

Although these performance enhancements may look different from what is
described above, it is still just an FL mapping at the end of the day- just
with more steps.


NOTE: The above is just to explain what is happening at a high level. Please
reference the implementation below to see how the BWT is actually currently
working
*/

const nullChar = "$"

// BWT Burrow Wheeler Transform
// Compresses and Indexes a given sequence so that it can be
// be used for search, alignment, and text extraction. This is
// useful for sequences so large that it would be beneficial
// to reduce its memory footprint while also maintaining a way
// to analyze and work with the sequence.
type BWT struct {
	// firstColumnSkipList is the first column of the BWT. It is
	// represented as a list of skipEntries because the first column of
	// the BWT is always lexicographically ordered. This saves time and memory.
	firstColumnSkipList []skipEntry
	// Column last column of the BWT- the actual textual representation
	// of the BWT.
	lastCoulmn waveletTree
	// suffixArray an array that allows us to map a position in the first
	// column to a position in the original sequence. This is needed to be
	// able to extract text from the BWT.
	suffixArray []int

	b      runInfo
	bPrime map[string]runInfo
	s      waveletTree
}

// Count represents the number of times the provided pattern
// shows up in the original sequence.
func (bwt BWT) Count(pattern string) (count int, err error) {
	defer bwtRecovery("Count", &err)

	searchRange := bwt.lfSearch(pattern)
	return searchRange.end - searchRange.start, nil
}

// Locate returns a list of offsets at which the begging
// of the provided pattern occurs in the original
// sequence.
func (bwt BWT) Locate(pattern string) (offsets []int, err error) {
	defer bwtRecovery("Locate", &err)

	searchRange := bwt.lfSearch(pattern)
	if searchRange.start >= searchRange.end {
		return nil, nil
	}

	numOfOffsets := searchRange.end - searchRange.start
	offsets = make([]int, numOfOffsets)
	for i := 0; i < numOfOffsets; i++ {
		offsets[i] = bwt.suffixArray[searchRange.start+i]
	}

	return offsets, nil
}

// Extract this allows us to extract parts of the original
// sequence from the BWT.
// start is the begging of the range of text to extract inclusive.
// end is the end of the range of text to extract exclusive.
// If either start or end are out of bounds, Extract will panic.
func (bwt BWT) Extract(start, end int) (extracted string, err error) {
	defer bwtRecovery("Extract", &err)

	if end > bwt.getLenOfOriginalStringWithNullChar()-1 {
		return "", fmt.Errorf("end [%d] exceeds the max range of the BWT [%d]", end, bwt.getLenOfOriginalStringWithNullChar()-1)
	}

	if start < 0 {
		return "", fmt.Errorf("start [%d] exceeds the min range of the BWT [0]", start)
	}

	strB := strings.Builder{}
	for i := start; i < end; i++ {
		fPos := bwt.getFCharPosFromOriginalSequenceCharPos(i)
		skip := bwt.lookupSkipByOffset(fPos)
		strB.WriteByte(skip.char)
	}

	return strB.String(), nil
}

// Len return the length of the sequence used to build the BWT
func (bwt BWT) Len() int {
	return bwt.getLenOfOriginalStringWithNullChar() - 1
}

// getFCharPosFromOriginalSequenceCharPos looks up mapping from the original position
// of the sequence to its corresponding position in the First Column of the BWT
func (bwt BWT) getFCharPosFromOriginalSequenceCharPos(originalPos int) int {
	for i := range bwt.suffixArray {
		if bwt.suffixArray[i] == originalPos {
			return i
		}
	}
	panic("Unable to find the corresponding original position for a character in the original sequence in the suffix array. This should not be possible and indicates a malformed BWT.")
}

// lfSearch LF Search- Last First Search.
// Finds the valid range within the BWT index where the provided pattern is possible.
// If the final range is <= 0, then the pattern does not exist in the original sequence.
func (bwt BWT) lfSearch(pattern string) interval {
	searchRange := interval{start: 0, end: bwt.getLenOfOriginalStringWithNullChar()}
	for i := 0; i < len(pattern); i++ {
		if searchRange.end-searchRange.start <= 0 {
			return interval{}
		}

		c := pattern[len(pattern)-1-i]
		skip, ok := bwt.lookupSkipByChar(c)
		if !ok {
			return interval{}
		}
		searchRange.start = skip.openEndedInterval.start + bwt.lastCoulmn.Rank(c, searchRange.start)
		searchRange.end = skip.openEndedInterval.start + bwt.lastCoulmn.Rank(c, searchRange.end)
	}
	return searchRange
}

// lookupSkipByChar looks up a skipEntry by its character in the First Column
func (bwt BWT) lookupSkipByChar(c byte) (entry skipEntry, ok bool) {
	for i := range bwt.firstColumnSkipList {
		if bwt.firstColumnSkipList[i].char == c {
			return bwt.firstColumnSkipList[i], true
		}
	}
	return skipEntry{}, false
}

// lookupSkipByOffset looks up a skipEntry based off of an
// offset of the Fist Column of the BWT.
func (bwt BWT) lookupSkipByOffset(offset int) skipEntry {
	if offset > bwt.getLenOfOriginalStringWithNullChar()-1 {
		msg := fmt.Sprintf("offset [%d] exceeds the max bound of the BWT [%d]", offset, bwt.getLenOfOriginalStringWithNullChar()-1)
		panic(msg)
	}
	if offset < 0 {
		msg := fmt.Sprintf("offset [%d] exceeds the min bound of the BWT [0]", offset)
		panic(msg)
	}

	for i := range bwt.firstColumnSkipList {
		if bwt.firstColumnSkipList[i].openEndedInterval.start <= offset && offset < bwt.firstColumnSkipList[i].openEndedInterval.end {
			return bwt.firstColumnSkipList[i]
		}
	}
	panic("figure out what to do here")
}

func (bwt BWT) getLenOfOriginalStringWithNullChar() int {
	return bwt.firstColumnSkipList[len(bwt.firstColumnSkipList)-1].openEndedInterval.end
}

type interval struct {
	start int
	end   int
}

type skipEntry struct {
	char byte
	// openEndedInterval start is inclusive and end is exclusive
	openEndedInterval interval
}

// New returns a BWT of the provided sequence
// The provided sequence must not contain the nullChar
// defined in this package. If it does, New will return
// an error.
func New(sequence string) (BWT, error) {
	if strings.Contains(sequence, nullChar) {
		return BWT{}, fmt.Errorf("Provided sequence contains the nullChar %s. BWT cannot be constructed", nullChar)
	}

	sequence += nullChar

	prefixArray := make([]string, len(sequence))
	for i := 0; i < len(sequence); i++ {
		prefixArray[i] = sequence[len(sequence)-i-1:]
	}

	sortPrefixArray(prefixArray)

	// OLD
	suffixArray := make([]int, len(sequence))
	lastColBuilder := strings.Builder{}

	// NEW
	charCount := 0
	sBuilder := strings.Builder{}
	var b runInfo
	bPrime := make(map[string]runInfo)

	var prevChar *byte
	for i := 0; i < len(prefixArray); i++ {
		currChar := sequence[getBWTIndex(len(sequence), len(prefixArray[i]))]
		if prevChar == nil {
			prevChar = &currChar
		}

		// NEW
		if currChar != *prevChar {
			sBuilder.WriteByte(*prevChar)
			b = append(b, i-charCount)
			addBPrimeEntry(bPrime, *prevChar, charCount)

			charCount = 0
			prevChar = &currChar
		}

		charCount++
		// OLD
		lastColBuilder.WriteByte(currChar)
		suffixArray[i] = len(sequence) - len(prefixArray[i])
	}
	sBuilder.WriteByte(*prevChar)
	b = append(b, len(prefixArray)-charCount)
	addBPrimeEntry(bPrime, *prevChar, charCount)

	fb := strings.Builder{}
	for i := 0; i < len(prefixArray); i++ {
		fb.WriteByte(prefixArray[i][0])
	}

	skipList := buildSkipList(prefixArray)

	// // TODO: do this more gracefully?
	// // Pad the end of b Prime so we can do max select queries with expected cumulative counts
	// for _, v := range skipList {
	// 	addBPrimeEntry(bPrime, v.char, 0)
	// }
	return BWT{
		firstColumnSkipList: skipList,
		lastCoulmn:          NewWaveletTreeFromString(lastColBuilder.String()),
		suffixArray:         suffixArray,

		s:      NewWaveletTreeFromString(sBuilder.String()),
		b:      b,
		bPrime: bPrime,
	}, nil
}

func addBPrimeEntry(bPrime map[string]runInfo, char byte, charCount int) {
	bPrimeOfChar, ok := bPrime[string(char)]
	if ok {
		bPrimeOfChar = append(bPrimeOfChar, charCount+bPrimeOfChar[len(bPrimeOfChar)-1])
	} else {
		bPrimeOfChar = runInfo{0, charCount}
	}
	bPrime[string(char)] = bPrimeOfChar
}

// buildSkipList compressed the First Column of the BWT into a skip list
func buildSkipList(prefixArray []string) []skipEntry {
	prevChar := prefixArray[0][0]
	skipList := []skipEntry{{char: prevChar, openEndedInterval: interval{start: 0}}}
	for i := 1; i < len(prefixArray); i++ {
		currChar := prefixArray[i][0]
		if currChar != prevChar {
			skipList[len(skipList)-1].openEndedInterval.end = i
			skipList = append(skipList, skipEntry{
				char:              currChar,
				openEndedInterval: interval{start: i},
			})
			prevChar = currChar
		}
	}
	skipList[len(skipList)-1].openEndedInterval.end = len(prefixArray)
	return skipList
}

// getBWTIndex returns the position of the character from the sequence used to build the BWT
// that corresponds the last character that would exist in the entry of the prefixArray that
// would be the last character if we were actually doing full rotations
func getBWTIndex(lenOfSequenceBeingBuilt, lenOfSuffixArrayVisited int) int {
	bwtCharIndex := lenOfSequenceBeingBuilt - lenOfSuffixArrayVisited - 1
	if bwtCharIndex == -1 {
		bwtCharIndex = lenOfSequenceBeingBuilt - 1
	}
	return bwtCharIndex
}

func sortPrefixArray(prefixArray []string) {
	slices.SortFunc(prefixArray, func(a, b string) bool {
		minLen := int(math.Min(float64(len(a)), float64(len(b))))
		for i := 0; i < minLen; i++ {
			if a[i] == b[i] {
				continue
			}
			if a[i] == nullChar[0] {
				return true
			}
			if b[i] == nullChar[0] {
				return false
			}
			return a[i] < b[i]
		}

		return len(a) < len(b)
	})
}

func bwtRecovery(operation string, err *error) {
	if r := recover(); r != nil {
		rErr := fmt.Errorf("BWT %s InternalError=%s", operation, r)
		*err = rErr
	}
}

// TODO: talk about how this should be a slice of ranks at their positions as sparse set bits
type runInfo []int

func (r runInfo) Select(rank int) int {
	return r[rank]
}

func (r runInfo) IsPastMaxRank(rank int) bool {
	return rank >= len(r)
}

func (r runInfo) Rank(startPos int) int {
	start := 0
	end := len(r) - 1
	for start < end {
		mid := start + (end-start)/2
		if r[mid] < startPos {
			start = mid + 1
			continue
		}
		if r[mid] > startPos {
			end = mid - 1
			continue
		}

		return mid
	}

	if r[start] > startPos {
		if start == 0 {
			return start
		}
		return start - 1
	}

	return start
}

// CountV2 represents the number of times the provided pattern
// shows up in the original sequence.
func (bwt BWT) CountV2(pattern string) (count int, err error) {
	defer bwtRecovery("Count", &err)

	searchRange := bwt.lfSearchV2(pattern)
	return searchRange.end - searchRange.start, nil
}

func (bwt BWT) lfSearchV2(pattern string) interval {
	searchRange := interval{start: 0, end: bwt.getLenOfOriginalStringWithNullChar()}
	for i := 0; i < len(pattern); i++ {
		if searchRange.end-searchRange.start <= 0 {
			return interval{}
		}

		c := pattern[len(pattern)-1-i]
		nextStart := bwt.getNextLfSearchOffset(c, searchRange.start)
		nextEnd := bwt.getNextLfSearchOffset(c, searchRange.end)
		searchRange.start = nextStart
		searchRange.end = nextEnd
	}
	return searchRange
}

func (bwt BWT) getNextLfSearchOffset(c byte, offset int) int {
	r := bwt.b.Rank(offset + 1)
	pv := bwt.s.Rank(c, r)

	s1, ok := bwt.lookupSkipByChar(c)
	if !ok {
		return 0
	}

	runInfoForChar, ok := bwt.bPrime[string(c)]
	if !ok {
		panic("better messaging no b prime")
	}

	s2 := runInfoForChar.Select(pv)

	cr := bwt.b.Rank(offset)
	cc := string(bwt.s.Access(cr))
	extraOffset := 0
	if c == cc[0] {
		o := bwt.b.Select(r)
		extraOffset += offset - o
	}

	return s1.openEndedInterval.start + s2 + extraOffset
}

func (bwt BWT) reconstructFColumn() string {
	str := ""
	for _, v := range bwt.firstColumnSkipList {
		for i := 0; i < v.openEndedInterval.end-v.openEndedInterval.start; i++ {
			str += string(v.char)
		}
	}
	return str
}
