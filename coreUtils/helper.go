package coreutils

func ToZero(byteSlice []byte) {
	for i := range byteSlice {
		byteSlice[i] = 0
	}
}