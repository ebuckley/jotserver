This is an authentication microservice, 

##Options

| param         | description   | type  |
| ------------- |:-------------:| -----:|
| connection    | broadcast port | string |
| db      | path to db file, will create if doesn't exist  |   string |
| statsd  | connection string for statsd server default is "127.0.0.1:8125"| string
	statsdConfig = flag.String("statsd", "127.0.0.1:8125", "statsd client location")


##Usage

```
jotserver
```
