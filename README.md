# gosync

`gosync` is a utility program written in golang that provides an S3 sync service running on OSX, linux and Windows.

Written in golang for the reason of cross compilable distributale binaries.

`gosync` does basic file watching using the `watcher` library (which means it doesn't have to use file system events yay!) and periodically performs the corresponding file sync events to S3 reported by `watcher`.

Configure a json file to list directories that should be monitored for any file changes.

`gosync` prioritises low footprint above redundancy through eventual consistency, planning ono making this configurable in the future. It also keeps a local sqlite cache to allow it to pick up wherever it left off in the case of a crash or failure.

The application is built with three main threads handling the process:

1) Thread handling any filewatch events from new files created, files moved, deleted or updated.

2) Thread doing validation on any new events found, like checks whether the file exists on s3 and does have the same MD5 hash.

3) This thread receives any tasks that require communication with AWS s3, like delete or put object calls.

Periodic status updates, reports and heartbeats can be sent by configuring the `webuiUrl` variable and providing the URL for each of the former events. It'll report this information via a POST request to the endpoints specified.

### Local File Cache

To make sure unneccesary uploads and move requests etc don't get sent to S3, a local cache is managed to store what the latest state of a certain file with a `canonicalKey` is, what/when it's last operation was and what the last known file MD5 hash that was synced.