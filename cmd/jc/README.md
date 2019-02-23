`jc`: Count JSON Records
========================

Count JSON records in the input stream.

Usage
-----
To get a simple count of the number of records in a stream:

```
> echo '{"f":"a"}{"f":"b"}' | jc --statistics
{
  "chunk-size": 1,
  "chunks": 2,
  "records": 2
}
```

The important here is `records` which is the total number of
JSON records in the stream.

Jc also keeps track of fixed sized chunks.  By default the `chunk-size`
is one, so the number of chunks (`chunks`) is the same as the number
of records.

It's easier to see with a larger data set, so first we'll generate 26
test records:

```
> for i in {a..z}; do echo "{\"f\":\"$i\"}"; done > /tmp/data.txt
> cat /tmp/data.txt
{"f":"a"}
...
{"f":"z"}
>
```

Now with the default chunk size of one:

```
> cat /tmp/data.txt | jc --statistics
{
  "chunk-size": 1,
  "chunks": 26,
  "records": 26
}> 
```

Now with a larger chunk size:

```
> cat /tmp/data.txt | jc --statistics --chunk-size 4
{
  "chunk-size": 4,
  "chunks": 6,
  "records": 26
}> 
```

Numbering Elements
------------------

The true magic comes with numbering elements.  Using the data file constructed
above:

```
> cat /tmp/data.txt | jc --chunk-size 4 
{
  "c": 0,
  "e": {
    "f": "a"
  },
  "i": 0,
  "s": 0
}{
  "c": 0,
  "e": {
    "f": "b"
  },
  "i": 1,
  "s": 1
}
...
}{
  "c": 6,
  "e": {
    "f": "z"
  },
  "i": 25,
  "s": 1
}>
```

Each record in the input stream is wrapped into a
dictionary.  The original record is attribute `e`. The
other attributes are:

  `i`: The records's index in the stream.
  `c`: The records's chunk number.
  `s`: The records's stride number.
  
The stride number is the element's count within its chunk.

The relation between `i`, `c`, and `s` is as follows:

```
  c = i div chunk-size
  s = i mod chunk-size
```


Splitting Into Files
---------------------
By combining `jc` and `jx` you can split a stream into multiple
files of no more than `chunk-size` elements:

```
> ls .
> cat /tmp/data.txt | jc | jx --ox chunk-{{c}}.json {{e}}
> ls .
chunk-0.json	chunk-1.json	chunk-2.json	chunk-3.json	chunk-4.json	chunk-5.json	chunk-6.json
> cat chunk-1.json
{"f":"e"}{"f":"f"}{"f":"g"}{"f":"h"}
>
