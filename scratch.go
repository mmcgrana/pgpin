package main

import (
	"fmt"
	api "github.com/mmcgrana/pgpin/pgpin-api"
)

func main() {
	api.DataStart()
	fmt.Println(api.DataTest())
}
