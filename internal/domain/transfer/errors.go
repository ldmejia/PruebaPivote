package transfer

import "errors"

var (
	ErrInvalidAmount    = errors.New("el monto debe ser mayor a cero")
	ErrSameAccount      = errors.New("la cuenta origen y destino no pueden ser la misma")
	ErrCurrencyMismatch = errors.New("las cuentas deben tener la misma moneda")

	ErrIdempotencyKeyRequired = errors.New("el header Idempotency-Key es requerido (maximo 100 caracteres)")
	ErrIdempotencyKeyReused   = errors.New("la Idempotency-Key ya se uso con otra solicitud distinta")
	ErrRequestInProgress      = errors.New("ya hay una solicitud en proceso con esa Idempotency-Key, intenta de nuevo")
	ErrDuplicatedKey          = errors.New("idempotency key duplicada")
	ErrTransferNotFound       = errors.New("transferencia no encontrada")

	ErrBeneficiaryRejected = errors.New("el core rechazo al beneficiario")
	ErrCoreUnavailable     = errors.New("el core no esta disponible")
)
