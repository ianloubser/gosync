package main

// type TaskOperation uint32

// const (
// Create TaskOperation = iota
// Delete
// )

type Event struct {
	event   TaskOperation
	message string
}
