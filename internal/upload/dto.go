package upload

// UploadResultOut — JSON returned after multipart save (filename + byte count).
type UploadResultOut struct {
	Filename string `json:"filename" example:"dashboard_screenshot_f47ac10b-58cc-4372-a567-0e02b2c3d479.png"`
	Bytes    int64  `json:"bytes" example:"1024"`
}
