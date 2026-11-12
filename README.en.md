# @go/log

Meta-driven structured logging with trace IDs, redaction and asynchronous output.

File output now defaults to independent gzip members plus one per-file `.dict` pool. Always move, archive and delete both files together. This is a versioned log format: use logv or the application's built-in log command; zcat is unsupported. `Config.NoCompress` retains legacy plaintext JSONL. Console and custom writers are unchanged.

Repeated JSON string tokens are interned on second occurrence, with bounded dictionaries. Dictionary additions are synced before data references. Complete members are checksummed; incomplete final members can be repaired when reopening. Only one writing process may own a file. Metadata is captured at file creation, so register models before first write and use a new file when changing layouts.

`server log` defaults to `-tail 100 -f`. Use `-tail N`, `-head N`, or `-cat`, optionally a filename before or after options. Explicit ranges exit unless `-f` is supplied. `-json` emits named JSON, `-raw` restored original JSONL, and `-color` enables ANSI. Output can be piped to native grep/awk. Configured follow tracks date rotation. Tail scans member headers and decompresses only selected blocks.

`logv [files ...]` restores compressed files and their automatically paired dictionaries. Without files it formats stdin; compressed stdin requires `-dict path`. Legacy plaintext remains readable.

Public query APIs: `QueryOptions`, `LogPath`, `DefaultConfig`, `Query`, `QueryFile`, `FormatStream`, `FormatArchiveStream`, and `RunCommand`. All query functions accept a context and an output writer. `FileWriter.Err` reports output failures; `WriterService.Stop` flushes, closes files and returns failures. See [Chinese API reference](README.md) for the complete logger/model/extension API and configuration.
