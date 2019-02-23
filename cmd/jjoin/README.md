JJoin: Relational Joins with JSON 
=================================

JJoin performs relational joins on streams of JSON objects.  This
allows you to treat streams of JSON obects similarly to relational
tables.

Join performs inner, outer, left-outer and right-outer along with
symmetric difference and difference.

Downloads
---------
You can get RPMs, DEBs, and OSX packages from [theblobshop.com](https://www.theblobshop.com/downloads/jjoin).

Usage
-----

Let's assume you have to files containing the following JSON:

`file1`:
```
{ "id": "L1", f1": "v1" }
{ "id": "L2", f1": "v2" }
```

`file2`:
```
{ "id": "R1", f1": "v1" }
{ "id": "R2", f1": "v3" }
```

The `inner` join returns:
```
> jjoin inner --left file1 --right file2 --using {{f1}}
{
  "left": {
    "f1": "v1",
    "id": "L1"
  },
  "right": {
    "f1": "v1",
    "id": "R1"
  }
}
```

Streams
-------
The joins use two files containing JSON streams: `--left` and `--right`. These streams are
either dictionaries or simple strings.  These can be shorted to `-l` and `-r`.

Separate Keys
-------------

You can choose separate keys for the left and the right.

`file1`:
```
{ "hostname": "foo.example.com", "ip": "172.17.31.20" }
```

`file2`:
```
{ "instance": "i78648798734", "private-ip": "172.17.31.20" }
```

```
> jjoin inner --left file1 --right file2 --left-key {{ip}} --right-key {{private-ip}}
{
  "left": {
    "hostname": "foo.example.com",
    "ip": "172.17.31.20"
  },
  "right": {
    "instance": "i78648798734",
    "private-ip": "172.17.31.20"
  }
}
```

Deeper Keys
-----------

Keys are a dot separated path of components from outer to inner.  The key
`a.b.c` would select `v1` from the dictionary:

```
{
  "a": {
     "b": {
        "c": {
          "v1"
        }
     }
  }
}
```

Simple Streams
--------------
The key `.` allows you to join with a string stream:
 
`file1`:
```
{ "hostname": "bar.example.com"}
{ "hostname": "baz.example.com"}
```

`file2`:
```
"foo.example.com"
"bar.example.com"
"baz.example.com"
```

```
> jjoin inner --left file1 --right file2 --left-key {{hostname}} --right-key {{.}}
{
  "left": {
    "hostname": "baz.example.com"
  }
  "right": "baz.examle.com"
}
{
  "left": {
    "hostname": "bar.example.com"
  }
  "right": "bar.examle.com"
}
```

Omitted Items
-------------
If the key does not exist within an object in the stream, then the
object is excluded from the join.
 
 
The Other Joins
---------------

All of these examples use the original files:

`file1`:
```
{ "id": "L1", f1": "v1" }
{ "id": "L2", f1": "v2" }
```

`file2`:
```
{ "id": "R1", f1": "v1" }
{ "id": "R2", f1": "v3" }
```

The `outer` join returns:
```
> jjoin outer --left file1 --right file2 --using f1
{
  "left": {
    "f1": "v1",
    "id": "L1"
  },
  "right": {
    "f1": "v1",
    "id": "R1"
  }
}
{
  "left": {
    "f1": "v2",
    "id": "L2"
  },
  "right": null
}
{
  "left": null,
  "right": {
    "f1": "v3",
    "id": "R2"
  }
}
```

The `left-outer` join returns:
```
> jjoin left-outer --left file1 --right file2 --using f1
{
  "left": {
    "f1": "v1",
    "id": "L1"
  },
  "right": {
    "f1": "v1",
    "id": "R1"
  }
}
{
  "left": {
    "f1": "v2",
    "id": "L2"
  },
  "right": null
}
```

The `right-outer` join returns:
```
> jjoin right-outer --left file1 --right file2 --using {{f1}}
{
  "left": {
    "f1": "v1",
    "id": "L1"
  },
  "right": {
    "f1": "v1",
    "id": "R1"
  }
}
{
  "left": null,
  "right": {
    "f1": "v3",
    "id": "R2"
  }
}
```

The `symm-diff` join returns:
```
> jjoin symm-diff --left file1 --right file2 --using {{f1}}
{
  "left": {
    "f1": "v2",
    "id": "L2"
  },
  "right": null
}
{
  "left": null,
  "right": {
    "f1": "v3",
    "id": "R2"
  }
}
```

The `subtract` join returns:
```
> jjoin subtract --left file1 --right file2 --using {{f1}}
{
    "f1": "v2",
    "id": "L2"
}
```

