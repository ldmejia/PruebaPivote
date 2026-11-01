package transfer

import "errors"

var (
	ErrInvalidAmount    = errors.New("el monto debe ser mayor a cero")
	ErrSameAccount      = errors.New("la cuenta origen y destino no pueden ser la misma")
	ErrCurrencyMismatch = errors.New("las cuentas deben tener la misma moneda")
)
