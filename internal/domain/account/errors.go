package account

import "errors"

var (
	ErrInvalidCurrency  = errors.New("moneda invalida, solo GTQ o USD")
	ErrInvalidBalance   = errors.New("el saldo inicial no puede ser negativo")
	ErrAccountNotFound  = errors.New("cuenta no encontrada")
	ErrDuplicatedNumber = errors.New("numero de cuenta duplicado")
	ErrInvalidDateRange = errors.New("rango de fechas invalido, usa from y to con formato YYYY-MM-DD y maximo 93 dias")
)
