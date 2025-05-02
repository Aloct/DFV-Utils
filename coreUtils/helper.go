package coreutils

import "strconv"

func ToZero(byteSlice []byte) {
	for i := range byteSlice {
		byteSlice[i] = 0
	}
}

func SixtyFourToDeci(rawNum int64) string {
	return strconv.FormatInt(rawNum, 10)
}