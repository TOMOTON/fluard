# Fluent Forward

This utility has been developed to test *Fluent Bit* inputs configured for 
the _Forward_ wire protocol. 

See also:

[The Fluent Bit Forward input plugin](https://docs.fluentbit.io/manual/pipeline/inputs/forward)

It should work with *Fluentd* as well, but this has not been tested.

See also:

[The Fluentd TCP Forward input plugin](https://docs.fluentd.org/input/forward)  
[The Fluentd UDP Forward input plugin](https://docs.fluentd.org/input/udp)  
[The Fluentd Unix Domain Socket Forward input plugin](https://docs.fluentd.org/input/unix)

Usage:

```
$ fluard [options] ADDRESS
```

## Options

Use -t or --tag to specify the Fluentd event tag

The default tag is `fluard.test`

Use -r or --record to specify the record as a JSON string or a JSON @\<file\>

The default record is:
```
{
    "message": "This is a test event",
	"local": {
		"user": "$USER",
		"host": "$HOST",
		"address": "$ADDRESS"
	}
}
```


## Installation

You can use the `go` tool to install `Fluent Forward`:

```
$ go install github.com/TOMOTON/fluard@latest
```


## Examples

Send the default record to a unix domain socket:

```
$ fluard unix:/run/fluent-bit.sock
```

Send a specific tag and record to a TCP endpoint:

```
$ fluard -t app.local -r '{"reason": "insufficient resources"}' tcp://remote.example.com:24224
```

