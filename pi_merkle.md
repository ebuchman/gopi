
Merkle Tree in Pi Calculus
--------------------------

# Byte Arrays

In pi calculus we express a length M byte array
as a pair of channels B and Bz. On B we broadcast M pairs of channels (N, Nz),
each pair expressing an integer (less than 256). 
We fire on Bz when we're done firing on B.

```
BYTE_ARRAY = B!(N1, N2, ... , Nm).Bz! | N1! | N2! | ... | Nm!
```

where each Nn is a pair of channels (Nn, Nnz) for the usual representation of integers.


# Hashing

We define the `HASH(B, Bh)` primitive such that it reads bytes off B and writes the digest onto Bh.

Now suppose we have two byte arrays running concurrently:

```
BA1 | BA2 
```

To concatenate them into BA1+BA2, we read both onto a new channel C:

```
Concat(BA1, BA2, C) = BA1?(b).C!(b).Concat(BA1, BA2) + BA1z?(z).Copy(BA2, C)
```

Note this is just the same as ADD. Now let's define

```
Concash(BA1, BA2, Ch) = Concat(BA1, BA2, C) | HASH(C, Ch)
```


# Merkle tree

The merklization of say four elements `{BA1, BA2, BA3, BA4}` looks like:

```
BA1 | BA2 | BA3 | BA4 | Concash(BA1, BA2, C1) | Concash(BA3, BA4, C2) | Concash(C1, C2, D)
```

We'd like a general expression that takes an input channel (BB) on which we send the byte arrays and a return channel (R) on which we can read off the root hash.

```
Merklize(BB, R) = (v BC) (v f) ( f! | merklize(BB, R, BC, f))

merklize(BB, R, BC, f) = (v g) BB?(b1).( BB?(b2).( Q(b1, b2, BC, f, g) | merklize(BB, R, BC, g) ) + BBz?.Copy(b1, R) ) + BBz?.( f?.BCz! | Merklize(BC, R) )

Q(b1, b2, c, f, g) = (v d) Concash(b1, b2, d) | f?.c!(d).g!
```

There are two loops: an inner loop over byte arrays, and an outer loop over heights in the merkle tree. 

For each instance of the outer loop, we will have a new channel (BC) on which to write (and then read) the next set of byte arrays (the concash's of the previous layer).

Each level we ascend in the outer loop leaves us with half as many byte arrays to process.

We would like to parallelize as much as possible, so that computations in all loops can occur concurrently, though processing in higher levels depends on completion of some of the processing in lower levels.

To ensure correct ordering despite the concurrency, an extra pair of channels (f, g) is used to make sure the hashes line up on the next channel (BC) in the right order.

The inner loop pulls on BB twice. If it gets two elements, it concashes them and continues. 

It it only finds one item, it fires it onto R (we're done).

If it finds no items, we've read all the byte arrays at this round, so go the next round of the outer loop.

Note we assume, for simplicity, that there are 2^n inputs, for some n.

# Reductions

Let's prove this reduces to the correct expression for some simple cases.

First, merklize two byte arrays:


```
BB!(BA1, BA2).BBz! | Merklize(BB, R) -->

(1) BB!(BA1, BA2).BBz! | (v f1)(v BC) ( f1! | (v g1) BB?(b1).( BB?(b2).( Q(b1, b2, BC, f1, g1) | merklize(BB, R, BC, g1) ) + BBz?.Copy(b1, R) ) + BBz?.( f1?.BCz! | Merklize(BC, R) ) )

(2) BBz! | (v f1)(v BC) ( f1! | (v g1) ( Q(BA1, BA2, BC, f1, g1) | merklize(BB, R, BC, g1) ) )

(3) BBz! | (v f1)(v BC) ( f1! | (v g1) ( (v d) Concash(BA1, BA2, d) | f1?.BC!(d).g1! | merklize(BB, R, BC, g1) ) )

(4) BBz! | (v BC) (v g1) ( (v d) Concash(BA1, BA2, d) | BC!(d).g1! | merklize(BB, R, BC, g1) ) 

(5) BBz! | (v BC) (v g1) ( (v d) Concash(BA1, BA2, d) | BC!(d).g1! | (v g2) BB?(b1).( BB?(b2).( Q(b1, b2, BC, g1, g2) | merklize(BB, R, BC, g2) ) + BBz?.Copy(b1, R) ) + BBz?.( g1?.BCz! | Merklize(BC, R) ) )

(6) (v BC) (v g1) ( (v d) Concash(BA1, BA2, d) | BC!(d).g1! |  g1?.BCz! | Merklize(BC, R) )

(7) (v BC1) (v g1) (v d) ( Concash(BA1, BA2, d) | BC1!(d).g1! |  g1?.BC1z! | Merklize(BC1, R) )

(8) (v BC1) (v g1) (v d) ( Concash(BA1, BA2, d) | BC1!(d).g1! |  g1?.BC1z! | (v f2) (v BC2) ( f2! | merklize(BC1, R, BC2, f2)) )

(9) (v BC1) (v g1) (v d) ( Concash(BA1, BA2, d) | BC1!(d).g1! |  g1?.BC1z! | (v f2) (v BC2) ( f2! | (v g3) BC1?(b1).( BC1?(b2).( Q(b1, b2, BC2, f2, g) | merklize(BC1, R, BC2, g) ) + BC1z?.Copy(b1, R) ) + BC1z?.( f2?.BC2z! | Merklize(BC2, R) ) ) )
	
(10) (v BC1) (v g1) (v d) ( Concash(BA1, BA2, d) | g1! |  g1?.BC1z! | (v f2) (v BC2) ( f2! | (v g3) ( BC1?(b2).( Q(d, b2, BC2, f2, g) | merklize(BC1, R, BC2, g) ) + BC1z?.Copy(d, R) ) 

(11) (v d) ( Concash(BA1, BA2, d) | Copy(d, R) )

QED
```

That one's easy because the outer loop only runs once. Let's do a merkle tree with four items:

```
BB!(BA1, BA2, BA3, BA4).BBz! | Merklize(BB, R) -->
```

The reduction is the same up to (5), tho instead of ( BBz! | P ) we have ( BB!(BA3, BA4).BBz! | P ) because we have two more items to send. 

So let's pickup at (5):

```
BB!(BA3, BA4).BBz! | (v BC) (v g1) (v d) ( Concash(BA1, BA2, d) | BC!(d).g1! | (v g2) BB?(b1).( BB?(b2).( Q(b1, b2, BC, g1, g2) | merklize(BB, R, BC, g2) ) + BBz?.Copy(b1, R) ) + BBz?.( g1?.BCz! | Merklize(BC, R) ) ) 

BBz! | (v BC) (v g1) (v d) ( Concash(BA1, BA2, d) | BC!(d).g1! | (v g2) ( Q(BA3, BA4, BC, g1, g2) | merklize(BB, R, BC, g2) ) ) 

BBz! | (v BC) (v g1) (v d1) ( Concash(BA1, BA2, d1) | BC!(d1).g1! | (v g2) (v d2) ( Concash(BA3, BA4, d2) | g1?.BC!(d2).g2! ) | merklize(BB, R, BC, g2) )  )

```

Now we've got the two concashes, and we're running the inner loop again. Since there's nothing left on BB, we move to the next round of the outer loop

```
BBz! | (v BC) (v g1) (v d1) ( Concash(BA1, BA2, d1) | BC!(d1).g1! | (v g2) (v d2) ( Concash(BA3, BA4, d2) | g1?.BC!(d2).g2!  | (v g3) BB?(b1).( BB?(b2).( Q(b1, b2, BC, g2, g3) | merklize(BB, R, BC, g3) ) + BBz?.Copy(b1, R) ) + BBz?.( g2?.BCz! | Merklize(BC, R) ) ) )

(v BC1) (v g1) (v d1) ( Concash(BA1, BA2, d1) | BC1!(d1).g1! | (v g2) (v d2) ( Concash(BA3, BA4, d2) | g1?.BC1!(d2).g2! | g2?.BC1z! | Merklize(BC1, R) ) )

(v BC1) (v g1) (v d1) ( Concash(BA1, BA2, d1) | BC1!(d1).g1! | (v g2) (v d2) ( Concash(BA3, BA4, d2) | g1?.BC1!(d2).g2! | g2?.BC1z! | (v f2) (v BC2) ( f2! | merklize(BC1, R, BC2, f2)) ) )

(v BC1) (v g1) (v d1) ( Concash(BA1, BA2, d1) | BC1!(d1).g1! | (v g2) (v d2) ( Concash(BA3, BA4, d2) | g1?.BC1!(d2).g2! | g2?.BC1z! | (v f2) (v BC2) ( f2! | (v g4) BC1?(b1).( BC1?(b2).( Q(b1, b2, BC2, f2, g4) | merklize(BC1, R, BC2, g4) ) + BC1z?.Copy(b1, R) ) + BC1z?.( f2?.BC2z! | Merklize(BC2, R) ) ) ) )

```

In the next outer round, we pull byte arrays off BC. We find two of them (d1, d2), and concash them

```
(v BC1) (v d1) (v d2) ( Concash(BA1, BA2, d1) | Concash(BA3, BA4, d2) | BC1z! | (v f2) (v BC2) ( f2! | (v g4) ( Q(d1, d2, BC2, f2, g4) | merklize(BC1, R, BC2, g4) ) ) )

(v BC1) (v d1) (v d2) ( Concash(BA1, BA2, d1) | Concash(BA3, BA4, d2) | BC1z! | (v BC2) (v g4) (v d3) ( Concash(d1, d2, d3) | BC2!(d3).g4! | merklize(BC1, R, BC2, g4) ) )

```

Now there's nothing left on BC1, so we go to the next round in the outer loop.

```
(v BC1) (v d1) (v d2) ( Concash(BA1, BA2, d1) | Concash(BA3, BA4, d2) | BC1z! | (v BC2) (v g4) (v d3) ( Concash(d1, d2, d3) | BC2!(d3).g4! | (v g5) BC1?(b1).( BC1?(b2).( Q(b1, b2, BC2, g4, g5) | merklize(BC1, R, BC2, g5) ) + BC1z?.Copy(b1, R) ) + BC1z?.( g4?.BC2z! | Merklize(BC2, R) ) ) )

(v BC2) (v d1) (v d2) (v d3) ( Concash(BA1, BA2, d1) | Concash(BA3, BA4, d2) | Concash(d1, d2, d3) | (v g4) ( BC2!(d3).g4! | g4?.BC2z! | Merklize(BC2, R) ) )

(v BC2) (v d1) (v d2) (v d3) ( Concash(BA1, BA2, d1) | Concash(BA3, BA4, d2) | Concash(d1, d2, d3) | (v g4) ( BC2!(d3).g4! | g4?.BC2z! | (v g6) (v BC3) ( g6! | merklize(BC2, R, BC3, g6)) ) )

(v BC2) (v d1) (v d2) (v d3) ( Concash(BA1, BA2, d1) | Concash(BA3, BA4, d2) | Concash(d1, d2, d3) | (v g4) ( BC2!(d3).g4! | g4?.BC2z! | (v g6) (v BC3) ( g6! | (v g7) BC2?(b1).( BC2?(b2).( Q(b1, b2, BC3, g6, g7) | merklize(BC2, R, BC3, g7) ) + BC2z?.Copy(b1, R) ) + BC2z?.( g6?.BC3z! | Merklize(BC3, R) ) ) ) )
```

In this outer loop, we're pulling byte arrays off BC2, but find there's only one, so we copy the final result (d3) onto R:

```
(v d1) (v d2) (v d3) ( Concash(BA1, BA2, d1) | Concash(BA3, BA4, d2) | Concash(d1, d2, d3) | Copy(d3, R) )

QED
```



# Merkle Proofs

Now that we can build static merkle trees in the pi calculus, let's adjust MERKLIZE to also output proof streams. We want something like:

```
MERKLE_PROOF(BB, BA, R, PSL, PSR)
```

where BB and R are as before (the data channel and a channel for the root hash). 
However, since we need to identify one of the byte arrays (say B, the one to prove), we have a second channel, BA, on which we fire B instead of firing it on BB.
Finally, PSL and PSR are channels on which to send the left and right proof nodes, respectively (our proof stream).
For simplicity, we let PS stand for (PSL, PSR).

We have to change the definition of BYTE_ARRAY to account for this:


```
BYTE_ARRAY(BB, BA) = (v f) (v g) BB!(N1, N2, ...).f!g?.BB!(... Nm) | f?.BA!(B).g!
```

This will enable us to have the equivalent of an if statement later on.

So now, we will pull on BB or BA. This will help us figure out what we need to fire onto the proof stream.


```
MerkleProof (BB, BA, R, PS) = (v BC) (v BCa) (v f) ( f! | merklize(BB, BA, R, PS, BC, BCa, f))

// we read off BB (element we're not trying to prove), BA (element we're trying to prove), or BBz (we've read them all)
merklize(BB, BA, R, PS, BC, BCa, f) = (v g) BB?(b1).QMC(BB, BA, R, PS, BC, BCa, b1, b2, f, g) + BA?(b1).QMCP(BB, BA, R, PS, BC, BCa, b1, b2, f, g) + BBz?.( f?.BCz! | MerkleProof(BC, BCa, R, PS) )

// Read off BB, fire on proof stream, QMa, or we're done and write to R
QMCP(BB, BA, R, PS, BC, BCa, b1, b2, f, g) = BB?(b2).( QMa(BB, BA, R, PS, BC, BCa, b1, b2, f, g) | PSR!(b2) ) + BBz?.Copy(b1, R)

// Read off BB and do QM, else read of BA and do QMa and fire on proof stream, else we're done and write to R
QMC(BB, BA, R, PS, BC, BCa, b1, b2, f, g) = BB?(b2).QM(BB, BA, R, PS, BC, BCa, b1, b2, f, g) + BA?(b2).( QMa(BB, BA, R, PS, BC, BCa, b1, b2, f, g) | PSL!(b1) ) + BBz?.Copy(b1, R)

// Q and merklize (inner loop) for a non proving node
QM(BB, BA, R, PS, BC, BCa, b1, b2, f, g) = Q(b1, b2, BC, f, g) | merklize(BB, BA, R, PS, BC, BCa, g)

// Q and merklize (inner loop) for a proving node (ie. fire on BCa instead of BC)
QMa(BB, BA, R, PS, BC, BCa, b1, b2, f, g) = Q(b1, b2, BCa, f, g) | merklize(BB, BA, R, PS, BC, BCa, g)

Q(b1, b2, c, f, g) = (v d) Concash(b1, b2, d) | f?.c!(d).g!
```

Now we can define:

```
MerkleVerify(PSL, PSR, B, R) = (v h) ( ( PSL!(b).Concash(b, B, h).f! + PSR!(b).Concash(B, b, h).f! + PSz!.Copy(h, R)) | f?.MerkleVerify(PSL, PSR, h, R) )
```

The output on R should be the same as when we ran MerkleProof.






