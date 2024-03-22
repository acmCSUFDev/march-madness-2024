# Intruder Alert!

In the heart of California State University, Fullerton, the ACM at CSUF is
buzzing with excitement during their annual event, "March Madness." As the
campus gears up for this thrilling extravaganza, an unexpected challenge arises.

The university's security system has detected mysterious access to one of its
buildings, and the ACM detectives are on the case! Your mission, should you
choose to accept it, is to sift through the access logs and identify the
intruders.

The scenario unfolds in the bustling month of February 2024, a time when the
campus is filled with the spirit of learning and innovation. As the ACM members,
you must utilize your coding prowess to decipher the logs and unveil the
trespassers' identities.

The access logs are your key to solving the mystery. Each log provides a
snapshot of individuals entering various buildings on campus. The logs are
timestamped, so you can determine the exact time of each entry.

Here is an example of the access logs:

```
1701617119: alice <- TG
1701578455: alice -> TG
1701489692: alice <- TG
1701572840: alice <- UP
1701435356: alice -> TG
1701470401: bob <- B
1701479255: alice -> TG
1701468315: bob -> B
1701535618: alice -> UP
```

Each line represents a single access log entry in the following format (each
variable is numbered):

```
${time}: ${person} -> ${building}
${time}: ${person} <- ${building}
1        2         3  4
```

1. The [Unix time](https://en.wikipedia.org/wiki/Unix_time) of the access log
   entry.
2. The name of the person who accessed the building.
3. The direction of the access (either `->` for entering or `<-` for exiting).
4. The name of the building that was accessed.

The access log entries may be in any order; you will need to analyze them to
determine the sequence of events.

In the example access logs above, we can sort the entries by timestamp:

```
1701435356: alice -> TG
1701468315: bob -> B
1701470401: bob <- B
1701479255: alice -> TG
1701489692: alice <- TG
1701535618: alice -> UP
1701572840: alice <- UP
1701578455: alice -> TG
1701617119: alice <- TG
```

Then, each entry can be interpreted as follows:

```
on December 1 at 12:55:56, Alice entered the building "TG"
on December 1 at 22:05:15, Bob entered the building "B"
on December 1 at 22:40:01, Bob exited the building "B"
on December 2 at 01:07:35, Alice entered the building "TG"
on December 2 at 04:01:32, Alice exited the building "TG"
on December 2 at 16:46:58, Alice entered the building "UP"
on December 3 at 03:07:20, Alice exited the building "UP"
on December 3 at 04:40:55, Alice entered the building "TG"
on December 3 at 15:25:19, Alice exited the building "TG"
```

## Part 1

Someone in our forensic team has observed a peculiar pattern in the access
logs: one of the entries is missing! They noted that someone somehow managed to
enter without later exiting and that they might've seen the person enter a
different building later on, even though they never exited the first building.
Your task is to identify the offending access log entry.

Using the example logs above, you can immediately spot that Alice entered the
building "TG" on December 1 at 12:55:56 then again on December 2 at 01:07:35
without ever leaving the building in between. The entrance log entry that
doesn't have a corresponding exit log entry is the suspicious access log entry,
and its Unix time corresponds to `1701435356`.

Based on this information, **what is the Unix time of the suspicious access log
entry** in the actual logs?

## Part 2

After further investigation, we've decided that this is enough evidence to
label any person who entered any building on that day as a suspect.

In the example logs above, the suspects are Alice and Bob who both entered at
least one building on December 1 (the day of the suspicious access log entry).
The total number of suspects is 2, and the number of suspicious entrances is 2,
therefore the answer is `4`.

To help us track down the suspects, **what is the number of suspects multiplied
by the number of suspicious entrances**?
