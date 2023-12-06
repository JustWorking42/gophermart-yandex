package apperrors

import "errors"

var ErrAlreadyRegisteredByThisUser = errors.New("order has already been uploaded by this user")

var ErrAlreadyRegisteredByAnotherUser = errors.New("order has already been uploaded by another user")

var ErrUserDoesNotExist = errors.New("user doesnt exist")

var ErrInsufficientBalance = errors.New("insufficient balance")
