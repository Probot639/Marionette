package util

// Chunk splits data into chunks of at most chunkSize bytes.
// Returns a slice of byte slices; the last chunk may be smaller.
func Chunk(data []byte, chunkSize int) [][]byte {
	if len(data) == 0 {
		return [][]byte{{}}
	}
	var chunks [][]byte
	for len(data) > 0 {
		end := chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[:end])
		data = data[end:]
	}
	return chunks
}

// Reassemble combines ordered chunks back into the original data.
// chunks must be in sequence order (0-indexed).
func Reassemble(chunks [][]byte) []byte {
	var total int
	for _, c := range chunks {
		total += len(c)
	}
	out := make([]byte, 0, total)
	for _, c := range chunks {
		out = append(out, c...)
	}
	return out
}
