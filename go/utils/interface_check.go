package utils

import "fmt"

type MyInterface interface {
	MyMethod()
}

type MyStruct struct { }

// This will throw an error if MyStruct does not implement MyInterface, without
// any runtime overhead. It's a compile-time check.
var _ MyInterface = (*MyStruct)(nil)

func (s *MyStruct) MyMethod() { // Now it won't throw an error because MyStruct implements MyInterface
	fmt.Println("MyMethod")
}