# lftpfetch

[![Build Status](https://travis-ci.org/martinp/lftpfetch.png)](https://travis-ci.org/martinp/lftpfetch)

A queue generator for [lftp](http://lftp.yar.ru).

## Usage

```
$ lftpfetch -h
Usage:
  lftpfetch [OPTIONS]

Application Options:
  -f, --config=FILE    Path to config (~/.lftpfetchrc)
  -n, --dryrun         Print generated command instead of running it
  -t, --test           Test and print parsed config
  -q, --quiet          Do not print actions

Help Options:
  -h, --help           Show this help message
```

## Example config

```json
{
  "Client": {
    "LftpPath": "lftp",
    "LftpGetCmd": "mirror"
  },
  "Sites": [
    {
      "Name": "foo",
      "Dir": "/dir",
      "LocalDir": "/tmp/{{ .Name }}/S{{ .Season }}/",
      "SkipSymlinks": true,
      "ParseTVShow": true,
      "MaxAge": "24h",
      "Patterns": [
        "^Dir1",
        "^Dir2"
      ],
      "Filters": [
        "(?i)incomplete"
      ]
    }
  ]
}
```