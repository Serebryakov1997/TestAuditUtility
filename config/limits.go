package config

import "errors"

const MaxBytes int64 = 10 << 20 // 10 MiB

var ErrConfigTooLarge = errors.New("configuration exceeds size limit")
