jtools: A Collection of JSON Tools
==================================

These are collection of tools to help working with json on the
command line. It includes the following tools:

`jc`: Count JSON objects on stdin.
`jjoin`: Perform simple relational joins on JSON objects.  
`jpar`: Run a command for each JSON object on stdin.  
`jx`: Expand a template for each JSON object.  
`l2j`: Convert plain text into JSON strings.


Downloads
---------
You can get RPMs, DEBs, and OSX packages from [theblobshop.com](https://www.theblobshop.com/downloads/jtools).


Examples
--------
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

