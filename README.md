```
        __                   ___     
  _____/ /_  ____  _________/ (_)___ 
 / ___/ __ \/ __ \/ ___/ __  / / __ \
/ /__/ / / / /_/ / /  / /_/ / / /_/ /
\___/_/ /_/\____/_/   \__,_/_/\____/ 
                                     
```

A distributed hash table (DHT) based on Chord

![build status](https://github.com/actions/kevinjqiu/chordio/workflows/.github/workflows/go.yml/badge.svg)


# Design

## Rank
In a chord system of rank `m`, a node is assigned an ID (using sha1) that's between `[0, 2**m)`.  e.g., in a chord ring of rank 7 (m=7), every node or key is hashed and assigned to an ID between `[0, 127)`.

## Node structure
In this implementation of chord (chordio), every node is a GRPC server maintaining a finger table of `m` entries.

### Basic Node operations
#### Find Successor
```
def find_succ(n, id)
    n' = find_pred(n, id)
    return n'.succ
```

#### Find Predecessor
```
def find_pred(n, id)
    n' = n
    while id not in (n', n'.succ]:
        n' = n'.closest_preceding_finger(id)
    return n'
```

#### Closest Preceding Finger

### Node join
When a node `n` is first started, it initiates its finger table like so:

```
def init_finger(n):
    for i in range(m):
        n.finger[i] = n
```

When a node is to join another node `n'`:
```
def join(n, n'):
    init_finger(n, n')
    update_others(n)
```

```
# initialize finger table of the local node
def init_finger(n, n'):
    succ = n.finger[0].node = n'.find_successor(n.finger[0].start)
    pred = succ.pred
    succ.pred = n
    for i in range(0, m-1):
        if n.finger[i+1].start in [n, n.finger[i].node):
            n.finger[i+1].node = n.finger[i].node
        else:
            n.finger[i+1].node = n'.find_successor(n.finger[i+1].start)
```

```
# update all nodes whose finger tables should refer to n
def update_others(n):
    for i in range(0, m):
        # find last node p whose ith finger might be n
        p = find_predecessor(n-2**i)
        p.update_finger(n, i)

# if s is ith finger of n, update n's finger table with s
def update_finger(n, s, i):
    if s in [n, n.finger[i].node):
        n.finger[i].node = s
        p = n.pred
        p.update_finger(s, i)
```
