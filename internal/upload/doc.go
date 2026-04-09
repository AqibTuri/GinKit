// Package upload handles multipart file uploads (size/extension checks, disk write).
// Uses internal/config for paths (UPLOAD_DIR) and limits; requires JWT on the route group.
// Stored files are named {sanitized_original_stem}_{uuid}{ext} under UploadDir.
package upload
