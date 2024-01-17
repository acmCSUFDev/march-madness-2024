# Intruder Alert!

In the heart of California State University, Fullerton, the ACM at CSUF is
buzzing with excitement during their annual event, "February Frenzy." As the
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
snapshot of individuals entering and exiting various buildings on campus. The
logs are timestamped, so you can determine the exact time of each entry.


Here is an example of the access logs:

```
alice -> CS [2023-10-01 08:00:00]
alice <- CS [2023-10-01 09:00:00]
bob <- E [2023-10-01 09:00:00]
bob -> E [2023-10-01 08:00:00]
charlie -> TSU [2023-10-01 08:00:00]
charlie <- TSU [2023-10-01 09:00:00]

suspects: alice, charlie
```

Each line represents a single access log entry. Within each entry:

- The first word is the name of the person who accessed the building.
- The second word is either `->` or `<-`, denoting whether the person entered or
  exited the building.
- The third word is enclosed in square brackets, and is the timestamp of the
  access log entry. It is formatted as `YYYY-MM-DD HH:MM:SS`.

The access log entries may be in any order. In the above example, the entry
where `bob` exited `CS` occurred before the entry where he entered
`E`.

# Part 1

We have managed to narrowed down a list of suspects, but we need more evidence
to identify the trespasser. **What is the total number of times that everyone
has exited the `CS` building?**

# Part 2

After further investigations, we've discovered that there were not just one, but
two trespassers! One trespasser was seen entering a building while another
trespasser was still inside. In other words, there are overlapping access logs
for any two different people. **What is the total number of overlapping access logs
for any two trespassers?**
