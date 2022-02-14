package model

type MessageStyle string

var (
	ErrorMessage   MessageStyle = "error"
	SuccessMessage MessageStyle = "success"
)

type (
	MessageFile struct {
		Url       string
		Extension string
	}

	Message struct {
		Title   string
		Message string
		Style   MessageStyle
	}
)
