#!/usr/bin/env python3

import json
import typing
from problems import problem_utils
from pydantic import BaseModel
import os

ITEMS_JSON = os.path.join(os.path.dirname(__file__), "items.json")
RECIPES_JSON = os.path.join(os.path.dirname(__file__), "recipes.json")


class ItemsJSON(BaseModel):
    relevant_items: list[str]
    miscellaneous_items: list[str]


class Problem(problem_utils.Problem):
    recipes: dict[str, list[str]]
    wanted: list[str]
    items: ItemsJSON

    def __init__(self, seed=0) -> None:
        super().__init__(seed)

        with open(ITEMS_JSON, "r") as f:
            self.items = ItemsJSON.model_validate_json(f.read())

        with open(RECIPES_JSON, "r") as f:
            self.recipes = json.loads(f.read())

        # Be lazy. I don't care!
        self.items.relevant_items = sanitize_names(self.items.relevant_items)
        self.items.miscellaneous_items = sanitize_names(self.items.miscellaneous_items)
        self.recipes = {
            sanitize_name(k): sanitize_names(v) for k, v in self.recipes.items()
        }

        self.wanted = self.rand.sample(self.items.relevant_items, 6)

    def generate_input(self, output: typing.IO | None = None):
        print(f'wanted: {", ".join(self.wanted)}', file=output)
        print("", file=output)
        for result, ingredients in self.recipes.items():
            print(f"{result} = {ingredients[0]} + {ingredients[1]}", file=output)

    def part1_answer(self) -> int:
        raise NotImplementedError
        return len(self.wanted)

    def part2_answer(self) -> int:
        raise NotImplementedError


def sanitize_name(name: str) -> str:
    name = name.lower()
    name = name.replace("+", "p")
    name = name.replace("=", "?")
    return name


def sanitize_names(names: list[str]) -> list[str]:
    return [sanitize_name(name) for name in names]


if __name__ == "__main__":
    problem_utils.main(Problem)
