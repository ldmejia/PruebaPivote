package account

import "errors"

var (
	ErrInvalidCurrency  = errors.New("moneda invalida, solo GTQ o USD")
	ErrInvalidBalance   = errors.New("el saldo inicial no puede ser negativo")
	ErrAccountNotFound  = errors.New("cuenta no encontrada")
	ErrDuplicatedNumber = errors.New("numero de cuenta duplicado")
)
