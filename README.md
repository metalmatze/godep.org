# godep.org [![Build Status](https://drone.github.matthiasloibl.com/api/badges/metalmatze/godep.org/status.svg)](https://drone.github.matthiasloibl.com/metalmatze/godep.org)

[![Docker Pulls](https://img.shields.io/docker/pulls/metalmatze/godep.org.svg?maxAge=604800)](https://hub.docker.com/r/metalmatze/godep.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/metalmatze/godep.org)](https://goreportcard.com/report/github.com/metalmatze/godep.org)


This is an experiment for a next generation [godoc.org](https://godoc.org).  
What if we add a lot of features to it, that are missing in its current form.

This project tries to shed light on this topic.

#### Why not simply improve GoDoc?

The purpose of this project is going beyond what GoDoc is currently capable of.
GoDoc uses Redis to store its data. This will not be sufficient for what we're planning to do.
Thus, right now, we use Postgres as Database.
Additionally we want to be able to experiment. If something works out really well,
I'm sure we can work on getting the feature into GoDoc as well.

## Development

Clone this repository:

```
git clone git@github.com:metalmatze/godep.git $GOPATH/src/github.com/metalmatze/godep.org
```


### Start Postgres

```
docker run -d -e POSTGRES_PASSWORD=postgres -p 5432:5432 --name godep-postgres postgres:10
```

Now you can run database migrations with [migrate](https://github.com/mattes/migrate/tree/master/cli#installation) 
which you need to install.

Run the migrations like from the root of this project:

```
migrate -database postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable -path migrations/ up
```


### Build godep.org

Get all dependencies. We use [golang/dep](https://github.com/golang/dep).  
Fetch all dependencies with:

```
dep ensure -v -vendor-only
```

Build the binary using `make`:

```
make install
```

In case you have `$GOPATH/bin` in your `$PATH` you can now simply start the bot by running:

```bash
GITHUB_TOKEN=XXX godep.org
```

_You obtain a token for GitHub here: [github.com/settings/tokens](https://github.com/settings/tokens)._
