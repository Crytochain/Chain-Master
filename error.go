package core
import "errors"
var (
	ErrKnownBlock = errors.New("block already known")
	ErrBlacklistedHash = errors.New("blacklisted hash")
	ErrNoGenesis = errors.New("genesis not found in chain")
)
var (
	ErrNonceTooLow = errors.New("nonce too low")
	ErrNonceTooHigh = errors.New("nonce too high")
	ErrGasLimitReached = errors.New("gas limit reached")
	ErrInsufficientFundsForTransfer = errors.New("insufficient funds for transfer")
	ErrInsufficientFunds = errors.New("insufficient funds for gas * price + value")
	ErrGasUintOverflow = errors.New("gas uint64 overflow")
	ErrIntrinsicGas = errors.New("intrinsic gas too low")
)
