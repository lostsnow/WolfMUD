package UID

type UID uint64

var Next chan UID

func init() {
	Next = make(chan UID)
	go func() {
		uid := UID(0)
		for {
			uid++
			Next <- uid
		}
	}()
}
