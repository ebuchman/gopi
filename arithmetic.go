package main

type Chan chan Chan

var C Chan

type Integer struct {
	n int  // cheating
	X Chan //successor channel
	Z Chan // zero channel
}

func NewInteger() *Integer {
	return &Integer{
		X: make(Chan),
		Z: make(Chan),
	}
}

func (i *Integer) FireX(c Chan) {
	i.X <- c
	i.n += 1
}

func (i *Integer) FireZ(c Chan) {
	i.Z <- c
}

func (i *Integer) PullX() Chan {
	i.n -= 1
	return i.X
}

func (i *Integer) PullZ() Chan {
	return i.Z
}

//----------------------------------------------------------------------------

// N represents a number using the pi calculus
// by firing on a channel N times, and then firing on the zero channel
// DEF: N = (x!)^n z!
func N(n int, i *Integer) {
	i.n = n
	for j := 0; j < n; j++ {
		i.FireX(C)
	}
	i.FireZ(C)
}

// Copy takes an integer i and copies it out to j
// DEF: Copy(i, j) = i.X?(a)Succ(i, j) + i.Z?(a)j.Z!(a)
func Copy(i, j *Integer) {
	select {
	case <-i.PullX():
		Succ(i, j)
	case <-i.PullZ():
		j.FireZ(C)
	}
}

// Succ takes an integer i and returns the successor j = i+1
// DEF: Succ(i, j) = j.X!(a)Copy(i, j)
func Succ(i, j *Integer) {
	j.FireX(C)
	Copy(i, j)
}

// Add takes two integers and adds them, returning the result on w
// DEF: Add(i, j, w) = i.X?(a)w.X!(a)Add(i, j, w) + i.Z?(a)Copy(j, w)
func Add(i, j, w *Integer) {
	select {
	case <-i.PullX():
		w.FireX(C)
		Add(i, j, w)
	case <-i.PullZ():
		Copy(j, w)
	}
}

// Double should copy i onto v and w
// DEF: Double(i, v, w) = (new i1, i2)[ Dup(i, i1, i2) | Copy(i1, v) | Copy(i2, w) ]
func Double(i, v, w *Integer) {
	i1, i2 := NewInteger(), NewInteger()
	go Dup(i, i1, i2)
	go Copy(i1, v)
	Copy(i2, w)
}

// Dup forwards i onto v and w as a precursor to being copied in Double.
// DEF: Dup(i, v, w) = i.X?(a)[ v.X!(a) | w.X!(a) | i.Z?(a)( v.Z!(a) | w.Z!(a) ) | Dup(i, v, w) ]
func Dup(i, v, w *Integer) {
	c := <-i.PullX()
	go func() {
		go v.FireX(c)
		w.FireX(c)
	}()
	go func() {
		z := <-i.PullZ()
		go v.FireZ(z)
		w.FireZ(z)
	}()
	Dup(i, v, w)
}

// AddTo adds the value i to w without firing on Z
// DEF: AddTo(i, w) = i.X?(a)w.X!(a)AddTo(i, w) + i.Z?(a)
func AddTo(i, w *Integer) {
	select {
	case <-i.PullX():
		w.FireX(C)
		AddTo(i, w)
	case <-i.PullZ():
	}
}

// Multiply computes the product of i and j and returns it on w
// DEF: Multiply(i, j, w) = j.X?(a)(new i1, i2)[ Double(i, i1, i2) | AddTo(i1, w) | Multiply(i2, j, w) ] + i.Z?(a)w.Z!(a)
func Multiply(i, j, w *Integer) {
	select {
	case <-j.PullX():
		i1, i2 := NewInteger(), NewInteger()
		go Double(i, i1, i2)
		go AddTo(i1, w)
		Multiply(i2, j, w)
	case <-j.PullZ():
		w.FireZ(C)
	}
}
