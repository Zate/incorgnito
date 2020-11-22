# ToDo Items

## Todo

- [X] Some kind of persistent storage for the Channel List that will persist between restarts
- [X] worker queue for the deltions inside the Purger struct
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
