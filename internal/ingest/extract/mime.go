package extract

// MIME type constants for supported document formats.
const (
	// Images (OCR).
	MimePNG  = "image/png"
	MimeJPEG = "image/jpeg"
	MimeTIFF = "image/tiff"
	MimeBMP  = "image/bmp"
	MimeGIF  = "image/gif"
	MimeWebP = "image/webp"

	// Documents.
	MimePDF      = "application/pdf"
	MimePlain    = "text/plain"
	MimeMarkdown = "text/markdown"
	MimeCSV      = "text/csv"
	MimeRTF      = "application/rtf"

	// Office Open XML.
	MimeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	MimeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	MimePPTX = "application/vnd.openxmlformats-officedocument.presentationml.presentation"

	// OpenDocument.
	MimeODT = "application/vnd.oasis.opendocument.text"
	MimeODS = "application/vnd.oasis.opendocument.spreadsheet"
	MimeODP = "application/vnd.oasis.opendocument.presentation"

	// Legacy Office.
	MimeDOC = "application/msword"
	MimeXLS = "application/vnd.ms-excel"
	MimePPT = "application/vnd.ms-powerpoint"
)
