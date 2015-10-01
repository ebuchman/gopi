package main

type Integer struct {
	X chan struct{} // successor channel
	Z chan struct{} // zero channel
}

func NewInteger() *Integer {
	return &Integer{
		X: make(chan struct{}),
		Z: make(chan struct{}),
	}
}

// N represents a number using the pi calculus
// by firing on a channel N times, and then firing on the zero channel
func N(n int, i *Integer) {
	for j := 0; j < n; j++ {
		i.X <- struct{}{}
	}
	i.Z <- struct{}{}
}

// Copy takes an integer i and copies it out to j
func Copy(i, j *Integer) {
	select {
	case <-i.X:
		Succ(i, j)
	case <-i.Z:
		j.Z <- struct{}{}
	}
}

// Succ takes an integer i and returns the successor j = i+1
func Succ(i, j *Integer) {
	j.X <- struct{}{}
	Copy(i, j)
}

// Add takes two integers and adds them, returning the result on w
func Add(i, j, w *Integer) {
	select {
	case <-i.X:
		w.X <- struct{}{}
		Add(i, j, w)
	case <-i.Z:
		Copy(j, w)
	}
}

// Double should copy i onto v and w
// XXX: the copies are blocking and happen in lockstep. this is bad
func Double(i, v, w *Integer) {
	select {
	case <-i.X:
		v.X <- struct{}{}
		w.X <- struct{}{}
		Double(i, v, w)
	case <-i.Z:
		v.Z <- struct{}{}
		w.Z <- struct{}{}
	}
}

// AddTo adds the value i to w without firing on Z
func AddTo(i, w *Integer) {
	select {
	case <-i.X:
		w.X <- struct{}{}
		AddTo(i, w)
	case <-i.Z:
	}
}

// Multiply computes the product of i and j and returns it on w
func Multiply(i, j, w *Integer) {
	select {
	case <-j.X:
		i1, i2 := NewInteger(), NewInteger()
		go Double(i, i1, i2)
		go AddTo(i1, w)
		Multiply(i2, j, w)
	case <-j.Z:
		w.Z <- struct{}{}
	}

}
