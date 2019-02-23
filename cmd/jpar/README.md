Jpar: Execute tasks in parallel
===============================

Yet another program to execute commands in parallel.  Jpar differs from others in
that it consumes the task descriptions from a stream of JSON objects, and it produces
output as a stream of JSON objects.  This make it easy to manipulate the output in
conjunction with other JSON tools.


Downloads
---------
You can get RPMs, DEBs, and OSX packages from [theblobshop.com](https://www.theblobshop.com/downloads/jpar).


Usage
-----
To execute `ls` in parallel over two directories: 

```
> echo '{"f":"/tmp"}{"f":"/usr"}' | jpar ls {{f}} 
{"cmd": ["ls", "/tmp"],"e":{"f":"/tmp"},"returncode":0,"stderr":"","stdout":"a\nb\n","outcome":"SUCCESS"}
{"cmd": ["ls", "/usr"],"e":{"f":"/usr"},"returncode":0,"stderr":"","stdout":"bin\nlib\n","outcome":"SUCCESS"}
>
```

For each JSON record in the input stream, one command is executed.  The arguments of this
command are expanded based on the JSON input.  Each command produces on JSON dictionary in
the output.  There output ordering is not related to the input ordering.

Any JSON objects can be used as input:
```
> echo '"/tmp""/usr"' | jpar ls {{.}} 
{"cmd": ["ls", "/tmp"],"e":"/tmp","returncode":0,"stderr":"","stdout":"a\nb\n","outcome":"SUCCESS"}
{"cmd": ["ls", "/usr"],"e":"/usr","returncode":0,"stderr":"","stdout":"bin\nlib\n","outcome":"SUCCESS"}
>
```

Command expansion is done with a mustache variant which is more closely documented in the
[jx tool](https://github.com/jmyounker/jx/blob/master/README.md).


You can specify environment variables with the `--env VAR=VALUE` option.
 
You can specify the execution directory with the `--dir DIRECTORY` option.

Both of these are expanded as templates using the input dictionary.


Result Field
-------------
If successful the output will contain the following fields:

* **cmd** An array containing the executed command.
* **e** The input entry.
* **returncode** The command's return code. An unexecuted command has returncode `-4242`.
* **stdout** Ihe command's stdout.
* **stderr** Ihe command's stderr.
* **outcome** Indicates if the command was executed correctly. Legal values are:
  * **SUCCESS** The command was executed to completion.
  * **FAILURE** The command could not be executed.

If a command fails do to an error in the execution there will additional fields:

* **error** An error message.

The following fields may also be defined:

* **env** A dictionary of environment variables and values.
* **dir** The directory from which the command was run.

The `--debug` flag adds the following fields to the output:

* **worker-id** An worker thread identifier.
* **prog** The path used to execute the command.

To Implement
------------
* Supply stdin
