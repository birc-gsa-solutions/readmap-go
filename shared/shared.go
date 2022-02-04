// You can create modules at this level and they will be
// interpreted as under module birc.au.dk, so to import
// package `shared` you need `import "birc.au.dk/gsa/shared"`

package shared

import "fmt"

func Preprocess(genome string) {
	fmt.Println("Preprocessing:", genome)
}

func Readmap(genome, reads string, dist int) {
	fmt.Println("Redmap genome", genome, "with", reads, "within distance", dist)
}
