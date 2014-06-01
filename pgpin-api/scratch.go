package main

import (
	"fmt"
)

func scratch() {
	fmt.Println(dataValidateNonempty("zero", ""))
	fmt.Println(dataValidateNonempty("blank", "  "))
	fmt.Println(dataValidateNonempty("valid", "abc"))
}
