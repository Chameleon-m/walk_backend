package redis

import "errors"

var (
	// ErrInvalidOnConnect ...
	ErrInvalidOnConnect = errors.New("OnConnect must not be nil")
	// ErrInvalidCredentialsProvider ...
	ErrInvalidCredentialsProvider = errors.New("CredentialsProvider must not be nil")
	// ErrInvalidMaxActiveConns ...
	ErrInvalidMaxActiveConns = errors.New("MaxActiveConns should be positive")
	// ErrInvalidTLSConfig ...
	ErrInvalidTLSConfig = errors.New("TLSConfig must not be nil")
	// ErrInvalidLimiter ...
	ErrInvalidLimiter = errors.New("limiter must not be nil")
)
