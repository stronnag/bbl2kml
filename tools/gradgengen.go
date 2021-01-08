package main

import (
	"fmt"
	"github.com/mazznoer/colorgrad"
)

func main() {

	var grad colorgrad.Gradient
	var gname []string = []string{"red", "rdylgn", "ylorrd"}

	fmt.Println(`package kmlgen

type GradSet struct  {
  R uint8
  G uint8
  B uint8
  A uint8
}

const (
	NUM_GRAD = 20
	GRAD_RED = 0
	GRAD_RGN = 1
	GRAD_YOR = 2
)

`)

	for j := 0; j < 3; j++ {
		switch j {
		case 1:
			grad = colorgrad.RdYlGn()
		case 2:
			grad = colorgrad.YlOrRd()
		default:
			grad = colorgrad.Reds()
		}

		fmt.Printf("var %s_grad []GradSet = []GradSet{\n", gname[j])

		for i := 0; i < 21; i++ {
			k := i
			if j&1 == 0 {
				k = 20 - i
			}
			c := grad.At(float64(k) / 20.0)
			r, b, g, a := c.RGBA()
			ur := uint8((r >> 8))
			ug := uint8((g >> 8))
			ub := uint8((b >> 8))
			ua := uint8((a >> 8))

			fmt.Printf("  { R: 0x%02x, G: 0x%02x, B: 0x%02x, A: 0x%02x },\n", ur, ub, ug, ua)
		}
		fmt.Println("}\n\n")
	}
	fmt.Println(`
func Get_gradset(idx int) ([]GradSet) {
	switch idx {
	case 		GRAD_YOR:
		return ylorrd_grad
	case GRAD_RGN:
		return rdylgn_grad
	default:
		return red_grad
	}
}
`)
}
