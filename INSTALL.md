Make sure your `$GOPATH` is exported:
```sh
$ export GOPATH=$HOME/go
```

Then type the following command:

```sh
$ go get -u github.com/shizeeg/jagod
```

`jagod` binary should be in `$GOPATH/bin` you may copy it whatever you'd like, for example to `/usr/local/bin`:
```sh
$ sudo mv $GOPATH/bin/jagod /usr/local/bin
```

Also, please edit `jagod.cfg` file and copy it to `/etc/jagod.cfg`:
```sh
$ $EDITOR $GOPATH/src/github.com/shizeeg/jagod/etc/jagod.cfg
```
```sh
$ sudo mv $GOPATH/src/github.com/shizeeg/jagod/etc/jagod.cfg /etc
```

Installation is now completed. It's time to give `jago` a try:
```sh
$ ./jagod
```
P.S.: I know it's a mess at the moment. I have plans to revive the project. Stay tuned. :)
