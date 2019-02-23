JX - a tool connecting JSON input to normal text
================================================

JX combines JSON input with a template to produce useful output.


Downloads
---------
You can get RPMs, DEBs, and OSX packages from [theblobshop.com](https://www.theblobshop.com/downloads/jx).


Usage
-----
At its simplest:

        > echo '{"a": "foo"}' | jx '{{a}}'
        foobar>

Multiple JSON objects result in multiple expansions:

        > echo '{"a": "foo"} {"a": "bar"}' | jx '{{a}}'
        foobar>

You can terminate each line with a newline using `-n`:

        > echo '{"a": "foo"} {"a": "bar"}' | jx -n '{{a}}'
        foo
        bar
        >

Text outside the mustaches is not left alone:

        > echo '{"a": "foo"} {"a": "bar"}' | jx -n 'this is {{a}}'
        this is foo
        this is bar
        >

You can also use arrays as JSON input:

        > echo '["foo"] ["bar"]' | jx -n 'this is {{1}}'
        this is foo
        this is bar
        >

With arrays the element index is the substitution variable:

        > echo '["foo", "bar"]' | jx -n 'index 1 is {{1}} and index 2 is {{2}}'
        index 1 is foo and index 2 is bar
        >

Names can refer to nested elements:

        > echo '{"foo":["bar1", "bar2"]}' | jx -n 'index foo.2 is {{foo.2}}'
        index foo.2 is bar2
        >


Simple JSON types are simple values:

        > echo '"foo" 42 true' | jx -n '{{.}}'
        foo
        42
        true
        >

Complex types are returned as JSON literals:

        > echo '{"foo":["bar1", "bar2"]}' | jx 'foo is {{foo}}'
        foo is ["bar1","bar2"]
        >

You can read the template from a file with the `-t` option:

        > echo 'this is {{a}}' > /tmp/tmpl
        > echo '{"a": "foo"} {"a": "bar"}' | jx -t /tmp/tmpl
        this is foo
        this is bar
        >

You can read the input from a file with the `-i` option:

        > echo {"a": "foo"} {"a": "bar"}' > /tmp/input
        > jx -n -i /tmp/input 'this is {{a}}'
        this is foo
        this is bar
        >

You can write output to a designated file with the `-o` option:


        > echo '{"a": "foo"} {"a": "bar"}' | jx -n -o /tmp/output 'this is {{a}}'
        > cat /tmp/output
        this is foo
        this is bar
        >

You can use a template to specify the location of the template using the `--tx` option:

        > echo 'template one is in file {{fn}}' > /tmp/t1
        > echo 'template two is in file {{fn}}' > /tmp/t2
        > echo '{"fn": "t1"} {"fn": "t2"}' | jx -n --tx /tmp/{{fn}}
        template one is in file t1
        template two is in file t2
        >

Similarly, you can use the `--ox` option to specify an output filename template:

        > echo '{"fn": "o1"} {"fn": "o2"}' | jx --ox /tmp/{{fn}} 'this is file {{fn}}'
        > cat /tmp/o1
        this is file o1
        > cat /tmp/o2
        this is file o2
        >

Note that by default the `--ox` option overwrites the previous contents of a file if it switches back:

        > echo '{"fn": "o1", "a": "first"} {"fn": "o2", "a": "second"} {"fn": "o1", "a": "third"}' | jx --ox /tmp/{{fn}} 'this was written {{a}}'
        > cat /tmp/o1
        this was written third
        > cat /tmp/o2
        this was written second
        >

The `-a` alters this behavior, appending instead of truncating:

        > echo '{"fn": "o1", "a": "first"} {"fn": "o2", "a": "second"} {"fn": "o1", "a": "third"}' | jx --ox /tmp/{{fn}} -a 'this was written {{a}}'
        > cat /tmp/o1
        this was written first
        this was written third
        > cat /tmp/o2
        this was written second
        >

The `-a` alters affects normal writes too:

        > echo "this was here to begin with" > /tmp/o1
        > echo '{"a": "new"} | jx -o /tmp/o1 -a 'this is {{a}}'
        > cat /tmp/o1
        this was here to begin with
        this is new
        >

Template Language
-----------------

By default JX uses modified [mustache](https://mustache.github.io/) templates.  It differs
from normal mustache templates in the following ways:

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

