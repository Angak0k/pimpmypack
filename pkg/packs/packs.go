package packs

import (
	"errors"
)

// ErrPackNotFound is returned when a pack is not found
var ErrPackNotFound = errors.New("pack not found")

// ErrPackContentNotFound is returned when no item are found in a given pack
var ErrPackContentNotFound = errors.New("pack content not found")
