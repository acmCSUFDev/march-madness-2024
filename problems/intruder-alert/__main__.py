#!/usr/bin/env python3

import math
import typing
import pydantic
from enum import Enum
from logging import debug
from problems import problem_utils
from datetime import (
    date as Date,
    datetime as DateTime,
    timezone as Timezone,
    timedelta as TimeDelta,
)
from .csuf_buildings import BUILDINGS


TIMEZONE = Timezone(TimeDelta(hours=-7))  # PDT
MIN_TIME = DateTime(2023, 12, 1, tzinfo=TIMEZONE)
PEOPLE = 200
ENTRIES = 10000

# Use for example
# PEOPLE = 2
# ENTRIES = 10


class Direction(Enum):
    ENTER = "IN"
    LEAVE = "OUT"

    def __str__(self) -> str:
        return "->" if self == Direction.ENTER else "<-"


class AccessEntry(pydantic.BaseModel):
    name: str
    time: DateTime
    building: str
    direction: Direction

    def __str__(self) -> str:
        time = self.time.strftime("%s")
        return f"{time}: {self.name} {self.direction} {self.building}"


class Problem(problem_utils.Problem):
    suspects: list[str] = []
    access_log: list[AccessEntry] = []
    crime_entry: tuple[AccessEntry, AccessEntry]

    def __init__(self, seed=0) -> None:
        super().__init__(seed)
        names = self.random_names(k=PEOPLE)

        times: dict[str, DateTime] = {}
        while len(self.access_log) < ENTRIES:
            name = self.rand.choice(names)
            min_time = times.get(
                name,
                MIN_TIME + TimeDelta(hours=self.rand.random() * 24),
            )

            enter_time = min_time
            leave_time = min_time + TimeDelta(hours=self.rand.random() * 24)
            times[name] = leave_time + TimeDelta(hours=self.rand.random() * 24)

            entrance = AccessEntry(
                name=name,
                time=enter_time,
                building=self.rand.choice(BUILDINGS),
                direction=Direction.ENTER,
            )

            exit = AccessEntry(
                name=name,
                time=leave_time,
                building=entrance.building,
                direction=Direction.LEAVE,
            )

            self.access_log.append(entrance)
            self.access_log.append(exit)

        # The accomplices will have the same entry times.
        # It will always be in December, but it can't overlap with existing entries.
        while True:
            crime_entry_ix = self.rand.randrange(0, len(self.access_log), 2)
            crime_entrance = self.access_log[crime_entry_ix]
            if crime_entrance.time.month == 12:
                break

        debug(f"Crime entry: {crime_entrance=}")

        # Delete the leave entry of the criminal.
        crime_exit = self.access_log[crime_entry_ix + 1]
        assert crime_exit.direction == Direction.LEAVE
        debug(f"Removing {crime_exit=}")
        self.access_log.pop(crime_entry_ix + 1)

        self.crime_entry = (crime_entrance, crime_exit)

    def random_time(
        self,
        min_time: DateTime,
        max_time: DateTime,
    ) -> DateTime:
        assert min_time < max_time
        min_unix = int(min_time.strftime("%s"))
        max_unix = int(max_time.strftime("%s"))
        # Randomize then convert back to a Date. This truncates the time.
        return DateTime.fromtimestamp(
            self.rand.randint(min_unix, max_unix), tz=TIMEZONE
        )

    def random_time_within_day(self, date: Date) -> DateTime:
        min_time = DateTime(date.year, date.month, date.day, tzinfo=TIMEZONE)
        max_time = min_time + TimeDelta(days=1)
        return self.random_time(min_time, max_time)

    def random_names(self, k: int) -> list[str]:
        return [
            generate_name(i) for i in self.rand.sample(range(NAME_MIN_I, NAME_MAX_I), k)
        ]

    def generate_input(self, output: typing.IO | None = None):
        for entry in self.rand.sample(self.access_log, len(self.access_log)):
            print(entry, file=output)

    """
    Part 1: There is an entrance entry that is missing its exit entry. What is
    the time of the entrance entry in unix time?

    Part 2: Everyone who entered on the same day as the criminal is a suspect.
    What's the number of entrances that happened on that day, multiplied by the
    number of suspects?
    """

    def part1_answer(self):
        return int(self.crime_entry[0].time.strftime("%s"))

    def part2_answer(self):
        crime_date = self.crime_entry[0].time.date()
        entrances = [
            entry
            for entry in self.access_log
            if entry.time.date() == crime_date and entry.direction == Direction.ENTER
        ]
        suspects = set(entry.name for entry in entrances)
        debug(f"{len(suspects)=} {len(entrances)=}")
        return len(entrances) * len(suspects)


VOWELS = "aeiou"
CONSONANTS = "bcdfghjklmnpqrstvwxyz"

# rule: cvcc
NAME_RULE = [CONSONANTS, VOWELS, CONSONANTS, CONSONANTS]
NAME_MIN_I = 0
NAME_MAX_I = math.prod([len(r) for r in NAME_RULE])
NAME_MULTIPLES = [
    # [v * c * c, c * c, c, 1]
    math.prod([len(r) for r in NAME_RULE[i + 1 :]])
    for i in range(0, len(NAME_RULE))
]


def generate_name(i: int) -> str:
    """
    Generate a name based on the index i. This is used to generate a unique name
    using random numbers.
    """
    assert NAME_MIN_I <= i < NAME_MAX_I
    return "".join(
        NAME_RULE[j][i // NAME_MULTIPLES[j] % len(NAME_RULE[j])]
        for j in range(0, len(NAME_RULE))
    )


if __name__ == "__main__":
    problem_utils.main(Problem)
