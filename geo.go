package main

import (
	"github.com/hongshibao/go-kdtree"
)

type EuclideanPoint struct {
	kdtree.Point
	Vec []float64
}

func (p EuclideanPoint) Dim() int {
	return len(p.Vec)
}

func (p EuclideanPoint) GetValue(dim int) float64 {
	return p.Vec[dim]
}

func (p EuclideanPoint) Distance(other kdtree.Point) float64 {
	var ret float64
	for i := 0; i < p.Dim(); i++ {
		tmp := p.GetValue(i) - other.GetValue(i)
		ret += tmp * tmp
	}
	return ret
}

func (p EuclideanPoint) PlaneDistance(val float64, dim int) float64 {
	tmp := p.GetValue(dim) - val
	return tmp * tmp
}

func NewEuclideanPoint(vals ...float64) *EuclideanPoint {
	ret := &EuclideanPoint{}
	for _, val := range vals {
		ret.Vec = append(ret.Vec, val)
	}
	return ret
}
