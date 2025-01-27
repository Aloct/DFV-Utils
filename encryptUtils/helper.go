package encryptUtils

func toZero(byteSlice []byte) {
	for i := range byteSlice {
		byteSlice[i] = 0
	}
}