package upload

// PresentResult builds the success payload for POST /upload.
func PresentResult(filename string, bytes int64) UploadResultOut {
	return UploadResultOut{Filename: filename, Bytes: bytes}
}
