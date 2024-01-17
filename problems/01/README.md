# Booting Up

February Frenzy is here! Most of our server infrastructure is up and running,
but we need to get the rest of it online before the event starts. We have a
bunch of services that need to be booted up, but we don't know which ones are
broken. We need you to help us figure out which services are broken so we can
fix them before the event starts.

We have a list of services that we need to boot up, as well as their status. The
status of each server is either `OK` or `STOP`. We need to boot up all of the
services that are `STOP`, but we don't want to waste time booting up services
that are already running.

As an example, our list might look like this:

```
[ OK ] redis
[STOP] ansible
[ OK ] caddy
[ OK ] rabbitmq
[STOP] postgres
```

In this case, we would need to boot up `ansible` and `postgres`, but not
`redis`, `caddy`, or `rabbitmq`.

## Part 1

Given a list of services and their statuses, **how many services do we need to
boot up?**

## Part 2

It seems that by starting the stopped services, we've caused ALL of our services
to crash! Now we have to restart ALL of them! **How many total restarts do we
need to do?**
