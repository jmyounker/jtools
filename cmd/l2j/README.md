`l2j`: Lines to JSON
==================

Converts lines of text into json strings.

Usage
-----
To convert a series of lines into a single string:

```
> printf "foo\nbar\nbaz\n" | l2j 
"foo\nbar\nbaz\n"
```

You can break the input every `n` lines using the `--stride` option:

```
> printf "foo\nbar\nbaz\nbam\n" | l2j --stride 2
"foo\nbar\n"
"baz\nbam\n"
```

To get every line as a separate string, use `--stride 1`:

```
> printf "foo\nbar\nbaz\nbam\n" | l2j --stride 1
"foo\n"
"bar\n"
"baz\n"
"bam\n"
```

You can break the input before each line matching a pattern:

```
> printf "foo\nbar\nbaz\n" | l2j --split-before bar
"foo\n"
"bar\nbaz\n"
```

You can break the input after each line matching pattern:

```
> printf "foo\nbar\nbaz\n" | l2j --split-after bar
"foo\nbar\n"
"baz\n"
```

Both are `--split-before` and `--split-after` take regular expressions:

```
> printf "foo\nbar\nbaz\n" | l2j --split-before 'b.r'
"foo\n"
"bar\nbaz\n"
```

```
> printf "foo\nbar\nbaz\n" | l2j --split-after 'b.r'
"foo\nbar\n"
"baz\n"
```


Each line includes its terminator, so regular expressions achored at the end
must match the line terminator, so the pattern `r$` does not match `bar\n`:

```
> printf "foo\nbar\nbaz\n" | l2j --split-after 'r$'
"foo\nbar\nbaz\n"
```

But the pattern `r\n$` does match `bar\n`:

```
> printf "foo\nbar\nbaz\n" | l2j --split-after 'r\n$'
"foo\nbar\n"
"baz\n"
```

Tricks
------

You can use `jc` to wrap lines into dictionaries:

```
> printf "foo\nbar\nbaz\n" | l2j --stride=1 | jc
{
  "c": 0,
  "e": "foo\n",
  "i": 0,
  "s": 0
}{
  "c": 1,
  "e": "bar\n",
  "i": 1,
  "s": 0
}{
  "c": 2,
  "e": "baz\n",
  "i": 2,
  "s": 0
}
```
 