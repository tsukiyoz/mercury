package errs

// code format - x xx xxx
// http resp code + module code + error code
// User module 01
const (
	UserInvalidInput      = 401001
	UserInvalidOrPassword = 401002
	UserDuplicateEmail    = 401003

	UserInternalServerError = 501001
)
