package apperrors

type ErrAlreadyRegisteredByThisUser struct{}

func (e *ErrAlreadyRegisteredByThisUser) Error() string {
	return "order has already been uploaded by this user"
}

type ErrAlreadyRegisteredByAnotherUser struct{}

func (e *ErrAlreadyRegisteredByAnotherUser) Error() string {
	return "order has already been uploaded by another user"
}

type ErrUserDoesNotExist struct{}

func (e *ErrUserDoesNotExist) Error() string {
	return "user does not exist"
}

type ErrInsufficientBalance struct{}

func (e *ErrInsufficientBalance) Error() string {
	return "insufficient balance"
}
