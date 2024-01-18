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
snapshot of individuals entering  various buildings on campus. The logs are
timestamped, so you can determine the exact time of each entry.


Here is an example of the access logs:

```
alice -> CS at 2023-10-02 08:00:00
bob -> E at 2023-10-01 08:00:00
charlie -> TSU at 2023-10-01 08:00:00

suspects: alice, charlie
```

Each line represents a single access log entry. Within each entry:

- The first variable is the name of the person who accessed the building.
- The second variable is the name of the building that was accessed.
- The third variable is enclosed in square brackets, and is the timestamp of the
  access log entry. It is formatted as `YYYY-MM-DD HH:MM:SS`.

The access log entries may be in any order. In the above example, the entry
where `alice` accessed the `CS` building is listed first, but it happened
after the entry where `bob` accessed the `E` building, even though the `bob`
entry occurred prior.

The logs only contain entries for the month of October, November, and December
2023. The logs do not contain entries for any other months.

## Part 1

We have received some new information about the case. According to an estimate
from the security system, the trespassing occurred around December. **What is
the total number of times that everyone has entered a building in December?**

## Part 2

After further investigations, we have discovered that there were not just one,
but multiple trespassers! In fact, we are fairly sure that a whole party broke
in! These people all entered different buildings, but they all entered at the
same time. **What is the total number of people who entered a building at the
same time multiplied by the number of access log entries that occurred at that
time?**
