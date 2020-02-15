# Beanstalk

Go client for [beanstalkd](http://beanstalkd.github.com/beanstalkd/).

## New home

**This package has a new home at
[https://github.com/beanstalkd/beanstalk](https://github.com/beanstalkd/beanstalk).**

The code is still available here to use,
for any clients that depend on this import path,
but there is no development on this repo.
Maintenance and development now happen
in the beanstalkd organization.

Please use `github.com/beanstalkd/beanstalk` in new code.

## Install

    $ go get github.com/kr/beanstalk

## Use

Produce jobs:

    c, err := beanstalk.Dial("tcp", "127.0.0.1:11300")
    id, err := c.Put([]byte("hello"), 1, 0, 120*time.Second)

Consume jobs:

    c, err := beanstalk.Dial("tcp", "127.0.0.1:11300")
    id, body, err := c.Reserve(5 * time.Second)
