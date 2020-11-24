# ToDo Items

## Todo

- [X] Some kind of persistent storage for the Channel List that will persist between restarts
- [X] worker queue for the deltions inside the Purger struct
- [ ] Big Item: Need to put in some logic to split out items into things that can be batched vs things that can't be batched and implement some logic to keep us under the API / Gateway Limits
`Clients are allowed 120 events every 60 seconds, meaning you can send on average at a rate of up to 2 events per second. Clients who surpass this limit are immediately disconnected from the Gateway, and similarly to the HTTP API, repeat offenders will have their API access revoked. Clients also have limit of concurrent Identify requests allowed per 5 seconds. If you hit this limit, the Gateway will respond with an Opcode 9 Invalid Session.`
- [ ] Needs some presence stuff
- [ ] write some docs?
- [ ] we should probably put some tests in this thing
- [ ] get it logging to a more detailed timestamp
- [ ] check it's memory usage profile and tune
- [ ] get docker-build and stack stuff working and deploy it
- [ ] Convert delete queue to go routine with waitgroup / semaphores to control max deletion workers
- [ ] expand !purge command to allow setting purge time limit (hrs) per channel
- [ ] expand !purge command to report status of all channels, queues and stats on messages deleted
- [ ] create a middleware to catch errors and a command to display errors across all commands etc.
- [ ] add a better logger that supports json and fields (logrus maybe)
- [ ] add !corgime to grab and display a random picture of a corgi.
- [ ] add command to grab and post news feeds on specific subjects
