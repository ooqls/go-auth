package common

func BytesToFloats(bytes []byte) []float32 {
	floats := make([]float32, len(bytes))
	for i := range bytes {
		floats[i] = float32(bytes[i])
	}
	return floats
}

func Float32ToBytes(floats []float32) []byte {
	bytes := make([]byte, len(floats))
	for i := range floats {
		bytes[i] = byte(floats[i])
	}
	return bytes
}
