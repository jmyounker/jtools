jtools: A Collection of JSON Tools
==================================

These are tools to help working with JSON on the command line. It
includes the following little programs:

* `jc`: Count JSON objects on stdin.
* `jjoin`: Perform simple relational joins on JSON objects.  
* `jpar`: Run a command for each JSON object on stdin.  
* `jx`: Expand a template for each JSON object.  
* `l2j`: Convert plain text into JSON strings.


Downloads
---------
You can get RPMs, DEBs, and OSX packages from [theblobshop.com](https://www.theblobshop.com/downloads/jtools).


Examples
--------

**Splitting document sections into files**

Suppose I have a document, `/tmp/f.txt`:
```
1
---
2
3
---
4
5
6
---
7
8
```

The `---` split the file into sections.  We can put each of these sections into a file:

```
> cat /tmp/f.txt | l2j --split-before `---` | jc | jx --output-expansion '/tmp/section.{{i}}' '{{e}}'
```

We can see the files created:
```
> ls /tmp/section.*
  /tmp/section.0	/tmp/section.1	/tmp/section.2	/tmp/section.3
```

And the contents of each file:

```
> cat /tmp/section.0
1
> cat /tmp/section.1
---
2
3
> cat /tmp/section.2
---
4
5
6
> cat /tmp/section.3
---
7
8
```

**Splitting document sections and counting lines**

Using the same file (`/tmp/f.txt`) we can count the lines in each
section.

```
> cat /tmp/f.txt | l2j --split-before '---' | jpar -i '{{.}}' -p 1 wc -l | jq
{
  "cmd": [
    "wc",
    "-l"
  ],
  "prog": "/usr/bin/wc",
  "returncode": 0,
  "stdout": "       1\n",
  "stderr": "",
  "outcome": "SUCCESS"
}
{
  "cmd": [
    "wc",
    "-l"
  ],
  "prog": "/usr/bin/wc",
  "returncode": 0,
  "stdout": "       3\n",
  "stderr": "",
  "outcome": "SUCCESS"
}
{
  "cmd": [
    "wc",
    "-l"
  ],
  "prog": "/usr/bin/wc",
  "returncode": 0,
  "stdout": "       4\n",
  "stderr": "",
  "outcome": "SUCCESS"
}
{
  "cmd": [
    "wc",
    "-l"
  ],
  "prog": "/usr/bin/wc",
  "returncode": 0,
  "stdout": "       3\n",
  "stderr": "",
  "outcome": "SUCCESS"
}
```

The `jpar` command runs a program once for each JSON object in the input. It
The `-i` option supplies a pattern for the programs standard input. The process's
standard in (generated), standard out, standard error, and return code are
recorded in the resulting json objects.

You can then use the tools of your choice (such as `jq` in the example above
to manipulate the output.) 

If the pattern `-i` template is not suppied and `.stdout` is defined in the input
dictionary then it stdout from there. This allows you to easily chain together
batches of commands.


Template System
---------------

By default the tools that perform expansions use a modified [mustache](https://mustache.github.io/) templates.
It differs from normal mustache templates in the following ways:

 * Simple expansions `{{x}}` do not perform HTML escaping.
 * Simple expansions `{{x}}` of JSON structures produce embedded JSON.
 * Unescaped expansions `{{{x}}}` behave exactly like normal expansions.

You can obtain strict mustache semantics with the `--strict-mustache` option:

        > echo '{"x": "&"}' | jx -n '{{x}}'
        &
        > echo '{"x": "&"}' | jx -n --strict-mustache '{{x}}'
        &amp;
        >

This includes the strange processing from the mustache library too:

        > echo '{"x": {"y": "z"}}' | jx -n '{{x}}'
        {"y":"z"}
        > echo '{"x": {"y": "z"}}' | jx -n --strict-mustache '{{x}}'
        map[y:z]
        >

Array lookup enhancements are still present when using `--strict-mustache`, so
maybe it's not so strictly mustache:

        > echo '["one", "two"]' | jx -n '{{2}}'
        two
        > echo '["one", "two"]' | jx -n '--strict-mustache {{2}}'
        two
        >

