// Command newid prints a new ID (Phase 4b decision 10) for a thing a builder
// is creating by hand, such as a character or record written into a world's
// data files. Each run is a new thing: never reuse an ID.
package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/ids"
)

func main() {
	cell := flag.String("cell", "0x0", "cell key of the room the thing is created in, hex (0x0 while rooms have no geography)")
	flag.Parse()
	c, err := strconv.ParseUint(*cell, 0, 64)
	if err != nil {
		fmt.Println("newid: -cell must be a number such as 0x0 or 0x89c25a31ec0f0e5f")
		return
	}
	fmt.Println(ids.New(c))
}
