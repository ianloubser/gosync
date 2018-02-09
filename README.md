# gosyncs3

`gosyncs3` is a utility program written in golang that provides an S3 sync service running on OSX, linux and Windows.

Written in golang for the reason of cross compilable distributale binaries.

`gosyncs3` does basic file watching using the `watcher` library (which means it doesn't have to use file system events yay!) and periodically performs the corresponding file sync events to S3 reported by `watcher`.

Configure a json file to list directories that should be monitored for any file changes.

`gosyncs3` prioritises low footprint above redundancy through eventual consistency, planning ono making this configurable in the future. It also keeps a local sqlite cache to allow it to pick up wherever it left off in the case of a crash or failure.

Periodic status updates, reports and heartbeats can be sent by configuring the `webuiUrl` variable and providing the URL for each of the former events. It'll report this information via a POST request to the endpoints specified.