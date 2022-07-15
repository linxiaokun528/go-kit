package tool

import "gopkg.in/fatih/set.v0"

func main() {
	a := set.New(set.NonThreadSafe)
	a.Add(1)
	a.Add(2)
}
