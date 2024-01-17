#!/usr/bin/env python3

import os
import enum
import random
import typing
import logging
import pydantic
import itertools
from lib import problem_utils
from datetime import datetime as DateTime, timedelta as TimeDelta


BUILDING_LIST_FILE = os.path.join(os.path.dirname(__file__), "csuf-buildings.txt")
MIN_TIME = DateTime(2023, 10, 1)
MAX_TIME = DateTime(2023, 12, 30, 23, 59, 59)
MIN_STAY_TIME = TimeDelta(hours=2)
MAX_STAY_TIME = TimeDelta(days=7)
PEOPLE = 500
ENTRIES = 50000


class AccessType(str, enum.Enum):
    ENTER = "->"
    LEAVE = "<-"

    def __str__(self) -> str:
        return self.value


class AccessEntry(pydantic.BaseModel):
    name: str
    building: str
    type: AccessType
    time: DateTime

    def __str__(self) -> str:
        time = self.time.strftime("%Y-%m-%d %H:%M:%S")
        return f"{self.name} {self.type} {self.building} [{time}]"


class Problem(problem_utils.Problem):
    accesses: dict[str, list[tuple[DateTime, DateTime]]] = {}
    suspects: list[str] = []
    access_log: list[AccessEntry] = []
    building_list = open(BUILDING_LIST_FILE, "r").read().splitlines()

    def __init__(self, seed=0) -> None:
        super().__init__(seed)

        for _ in range(PEOPLE):
            while True:
                name = generate_name(self.rand)
                if name not in self.accesses:
                    break
            self.accesses[name] = []

        names = self.names()

        # The accomplices will have the same entry times.
        # It will always be in December.
        crime_time_enter = generate_time(
            self.rand,
            min_time=DateTime(2023, 12, 1),
            max_time=MAX_TIME - MAX_STAY_TIME,
        )
        crime_time_leave = generate_time(
            self.rand,
            min_time=crime_time_enter + MIN_STAY_TIME,
            max_time=crime_time_enter + MAX_STAY_TIME,
        )
        crime_time = (crime_time_enter, crime_time_leave)
        accomplices = self.rand.sample(names, k=self.rand.randint(6, 12))

        for _ in range(ENTRIES // 2):
            name = self.rand.choice(names)

            if name in accomplices and self.coin_flip(0.25):
                times = crime_time
            else:
                enter_time = generate_time(self.rand, max_time=MAX_TIME - MAX_STAY_TIME)
                leave_time = generate_time(
                    self.rand,
                    min_time=enter_time + MIN_STAY_TIME,
                    max_time=enter_time + MAX_STAY_TIME,
                )
                times = (enter_time, leave_time)

            building = self.rand.choice(self.building_list)

            enter = AccessEntry(
                name=name,
                building=building,
                type=AccessType.ENTER,
                time=times[0],
            )
            leave = AccessEntry(
                name=name,
                building=enter.building,
                type=AccessType.LEAVE,
                time=times[1],
            )

            self.access_log.append(enter)
            self.access_log.append(leave)
            self.accesses[name].append((enter.time, leave.time))

    def names(self) -> list[str]:
        return list(self.accesses.keys())

    def generate_input(self, output: typing.IO | None = None):
        for entry in self.access_log:
            print(entry, file=output)

    def part1_answer(self):
        return len(
            list(
                filter(
                    lambda e: e.time.month == 12 and e.type == AccessType.ENTER,
                    self.access_log,
                )
            )
        )

    def part2_answer(self):
        def slow_solution():
            encounters: list[AccessEntry] = []
            accomplices: list[str] = []
            for i in range(len(self.access_log)):
                a = self.access_log[i]
                is_crime = False
                for j in range(len(self.access_log)):
                    if i == j:
                        continue
                    b = self.access_log[j]
                    if a.type == b.type and a.time == b.time:
                        is_crime = True
                        break
                if is_crime:
                    encounters.append(a)
                    if a.name not in accomplices:
                        accomplices.append(a.name)
            return len(encounters) * len(accomplices)

        def fast_solution():
            log_set: dict[str, list[AccessEntry]] = {}
            for entry in self.access_log:
                key = f"{entry.type} {entry.time}"
                encountered = log_set.get(key, [])
                encountered.append(entry)
                log_set[key] = encountered

            counts = sorted(
                [len(entries) for entries in log_set.values()],
                reverse=True,
            )
            colluded_items = counts[:2]
            collided_entries = [
                entry
                for entries in log_set.values()
                if len(entries) in colluded_items
                for entry in entries
            ]

            logging.debug("\n".join([str(e) for e in collided_entries]))
            accomplices = set([entries.name for entries in collided_entries])
            return len(collided_entries) * len(accomplices)

        # return problem_utils.run_fast_slow(fast_solution, slow_solution)
        return fast_solution()


def generate_name(rand: random.Random) -> str:
    VOWELS = "aeiou"
    CONSONANTS = "bcdfghjklmnpqrstvwxyz"
    return "".join(
        rand.choices(CONSONANTS, k=1)
        + rand.choices(VOWELS, k=1)
        + rand.choices(CONSONANTS, k=2)
    )


def generate_time(
    rand: random.Random,
    min_time: DateTime | None = None,
    max_time: DateTime | None = None,
) -> DateTime:
    min_time = max(min_time, MIN_TIME) if min_time else MIN_TIME
    max_time = min(max_time, MAX_TIME) if max_time else MAX_TIME

    assert min_time < max_time
    assert min_time >= MIN_TIME
    assert max_time <= MAX_TIME

    return DateTime.fromtimestamp(
        rand.randint(
            round(min_time.timestamp()),
            round(max_time.timestamp()),
        )
    )


if __name__ == "__main__":
    problem_utils.main(Problem)
