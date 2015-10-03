package main

type Chan chan Chan

var C Chan

type Integer struct {
	X Chan //successor channel
	Z Chan // zero channel

	n int // cheating (debuging)
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
// DEF: Copy(i, j) = iX?(a)Succ(i, j) + iZ?(a).jZ!(a)
func Copy(i, j *Integer) {
	select {
	case <-i.PullX():
		Succ(i, j)
	case <-i.PullZ():
		j.FireZ(C)
	}
}

// Succ takes an integer i and returns the successor j = i+1
// DEF: Succ(i, j) = jX!(a).Copy(i, j)
func Succ(i, j *Integer) {
	j.FireX(C)
	Copy(i, j)
}

// Add takes two integers and adds them, returning the result on w
// DEF: Add(i, j, w) = iX?(a).wX!(a).Add(i, j, w) + iZ?(a).Copy(j, w)
func Add(i, j, w *Integer) {
	select {
	case <-i.PullX():
		w.FireX(C)
		Add(i, j, w)
	case <-i.PullZ():
		Copy(j, w)
	}
}

// Subtract j from i and return onto w. i must be larger than j.
// DEF: Subtract(i, j, w) = Sub(i, j).Copy(i, w)
func Subtract(i, j, w *Integer) {
	Sub(i, j)
	Copy(i, w)
}

// Takes the absolute value of the difference between i and j and copies onto w
// DEF: Diff(i, j, w) = (new d)( sub(i, j, d) | sub(j, i, d) | d?(a).Succ(a, w)
func Diff(i, j, w *Integer) {
	d := make(chan *Integer)
	go sub(i, j, d)
	go sub(j, i, d)
	a := <-d
	Succ(a, w)
}

// Subtract j from i. i must be larger than j. broadcast i back on d
// NOTE d is a channel that takes an int, so its really a channel on which we pass two channels ...
// DEF: sub(i, j, d) = jX?(a).sub(i, j, d) + jZ?(a).d!(i)
func sub(i, j *Integer, d chan *Integer) {
	select {
	case <-j.PullX():
		<-i.PullX()
		sub(i, j, d)
	case <-j.PullZ():
		d <- i
	}
}

// Subtract j from i. i must be larger than j
// DEF: Sub(i, j) = jX?(a).Sub(i, j) + jZ?(a)
func Sub(i, j *Integer) {
	select {
	case <-j.PullX():
		<-i.PullX()
		Sub(i, j)
	case <-j.PullZ():
	}
}

// Double should copy i onto v and w
// DEF: Double(i, v, w) = (new i1, i2)[ Dup(i, i1, i2) | Copy(i1, v) | Copy(i2, w) ]
func Double(i, v, w *Integer) {
	i1, i2 := NewInteger(), NewInteger()
	go Dup(i, i1, i2)
	go func() {
		Copy(i1, v)
	}()
	go func() {
		Copy(i2, w)
	}()
}

// Dup forwards i onto v and w as a precursor to being copied in Double.
// DEF: Dup(i, v, w) = iX?(a)[ vX!(a).d!(c)| wX!(a).d!(c) | d?(a).d?(a).Dup(i, v, w) ] + iZ?(a)( vZ!(a) | wZ!(a) )
func Dup(i, v, w *Integer) {
	d := make(Chan)
	select {
	case c := <-i.PullX():
		go func() {
			v.FireX(c)
			d <- C
		}()
		go func() {
			w.FireX(c)
			d <- C
		}()
		_, _ = <-d, <-d
		Dup(i, v, w)
	case z := <-i.PullZ():
		go v.FireZ(z)
		w.FireZ(z)
	}
}

// AddTo adds the value i to w without firing on Z
// DEF: AddTo(i, w) = iX?(a).wX!(a).AddTo(i, w) + iZ?(a)
func AddTo(i, w *Integer) {
	select {
	case <-i.PullX():
		w.FireX(C)
		AddTo(i, w)
	case <-i.PullZ():
	}
}

// Multiply computes the product of i and j and returns it on w
// DEF: Multiply(i, j, w) = jX?(a).(new i1, i2)[ Double(i, i1, i2) | AddTo(i2, w) | Multiply(i1, j, w) ] + iZ?(a).FlushFire(i, w)
func Multiply(i, j, w *Integer) {
	select {
	case <-j.PullX():
		i1, i2 := NewInteger(), NewInteger()
		go Double(i, i1, i2)
		go AddTo(i2, w)
		Multiply(i1, j, w)
	case <-j.PullZ():
		FlushFire(i, w)
	}
}

// Flush out an integer and fire the zero on another
// FlushFire(i, w) = x?(a).FlushFire(i, w) + z?(a).w!(a)
func FlushFire(i, w *Integer) {
	select {
	case <-i.X:
		FlushFire(i, w)
	case <-i.Z:
		w.FireZ(C)
	}
}
