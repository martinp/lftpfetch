# lftpq

[![Build Status](https://travis-ci.org/martinp/lftpq.svg)](https://travis-ci.org/martinp/lftpq)

A queue generator for [lftp](http://lftp.yar.ru).

## Usage

```
# lftpq -h
Usage:
  lftpq [OPTIONS]

Application Options:
  -f, --config=FILE           Path to config (default: ~/.lftpqrc)
  -n, --dryrun                Print queue and exit
  -F, --format=[lftp|json]    Format to use in dryrun mode (default: lftp)
  -t, --test                  Test and print config
  -q, --quiet                 Do not print output from lftp
  -i, --import=SITE           Read remote paths from stdin and build a queue for SITE

Help Options:
  -h, --help                  Show this help message
```

## Example config

```json
{
  "Default": {
    "Client": {
      "Path": "lftp",
      "GetCmd": "mirror"
    }
  },
  "Sites": [
    {
      "Name": "foo",
      "Dir": "/dir",
      "LocalDir": "/tmp/{{ .Name }}/S{{ .Season }}/",
      "SkipSymlinks": true,
      "SkipExisting": true,
      "SkipFiles": true,
      "Merge": true,
      "Priorities": [
        "^important",
        "^less\\.important"
      ],
      "Parser": "show",
      "MaxAge": "24h",
      "Patterns": [
        "^Dir1",
        "^Dir2"
      ],
      "Filters": [
        "(?i)incomplete"
      ],
      "Replacements": [
        {
          "Pattern": "\\.The\\.",
          "Replacement": ".the."
        }
      ],
      "PostCommand": "/usr/local/bin/post-process.sh"
    }
  ]
}
```

## Configuration options

`Default` holds the default site configuration, which will apply to all sites.
All options can be overridden per site. This is useful when you want to apply
the same options to multiple sites.

`Client` holds configuration related to lftp.

`Path` sets the path to the lftp executable (if only the base name is given,
`PATH` will be used for lookup).

`GetCmd` sets the lftp command to use when downloading, this can also be an
alias. For example: If you have `alias m "mirror --only-missing"` in your
`.lftprc`, then `LftpGetCmd` can be set to `m`.

`Sites` holds the configuration for each individual site.

`Name` is the bookmark or URL of the site. This is passed to the `open` command in lftp.

`Dir` is the remote directory used to generate the queue.

`LocalDir` is the local directory where files should be downloaded. This can be
a template. When the `show` parser is used, the following template variables are
available:

Variable  | Description            | Type   | Example
--------- | -----------------------|------- | -------
`Name`    | Name of the show       | string | `The.Wire`
`Season`  | Show season            | int    | `1`
`Episode` | Show episode           | int    | `5`
`Release` | Release/directory name | string | `The.Wire.S01E05.720p.BluRay.X264`

When using the `movie` parser, the following variables are available:

Variable  | Description            | Type   | Example
--------- | -----------------------| -------| -------
`Name`    | Movie name             | string | `Apocalypse.Now`
`Year`    | Production year        | int    | `1979`
`Release` | Release/directory name | string | `Apocalypse.Now.1979.720p.BluRay.X264`

All variables can be formatted with `Sprintf`. For example `/mydir/{{ .Name
}}/S{{ .Season | Sprintf "%02" }}/` would format the season using two decimals
and would result in `/mydir/The.Wire/S01`.

`SkipSymlinks` determines whether to ignore symlinks when generating the queue.

`SkipExisting` determines whether to ignore non-empty directories that already
exist in `LocalDir`.

`SkipFiles` determines whether to ignore files when generating the queue. When
`true` only directories will be included in the queue. Files inside a directory
will still be transferred.

`Merge` determines whether files/directories that exist locally should be merged
into the queue before deduplication takes place. This information can then be
used when doing post-processing of the queue.

For example: Candidate `A` is transferred during session `1` and candidate `B`
is transferred during session `2`. If candidate `B` has a higher priority than
`A`, so that `A` is considered a duplicate of `B`, `B` will be transferred
during session `2`. This means that both candidates exist on disk after session
`2` ends. If `Merge` is `true`, the queue passed to `PostCommand` will contain
candidate `A` along with its duplication status (the fields `Merged` and
`Duplicate` will both be `true`).

`Priorities` is a list of patterns used to deduplicate directories which contain
the same media (e.g. same show, season and episode, but different release).
Directories are deduplicated based on the order of matching patterns, where the
earliest match is given the highest weight. For example, if the items
`Foo.1.important` and `Foo.1.less.important` are determined to be the same
media, then given the priorities in the example above, `Foo.1.important` would
be kept and `Foo.2.less.important` would be removed from the queue.

`Parser` sets the parser to use when parsing media. Valid values are `show`,
`movie` or empty string (disable parsing).

`MaxAge` sets the maximum age of directories to consider for the queue. If a
directory is older than `MaxAge`, it will always be excluded. `MaxAge` has
precedence over `Patterns` and `Filters`.

`Patterns` is a list of patterns (regular expressions) used when including
directories. A directory matching any of these patterns will be included in the
queue.

`Filters` is a list of patterns used when excluding directories. A directory
matching any of these patterns will be excluded from the queue. `Filters` has
precedence over `Patterns`.

`Replacements` is a list of replacements that can be used to replace
misspellings or incorrect casing in media titles. `Pattern` is a regular
expression and `Replacement` is the replacement string. If multiple replacements
are given, they will be processed in the order they are listed.

`PostCommand` specifies a command for post-processing of the queue. The queue
will be passed to the command on stdin, in JSON format. Leave empty to disable.
