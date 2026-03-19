package model

type Key string
type Val string

type KeyVal struct {
	Key Key `json:"key"`
	Val Val `json:"val"`
}
