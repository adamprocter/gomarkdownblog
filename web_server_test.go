package main

import "testing"

func Test(t *testing.T) {
	//	main()
	a, err := Posts()
	t.Log(a, err)

}
