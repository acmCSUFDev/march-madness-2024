#!/usr/bin/env python3

import json
import typing
from logging import debug
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
            self.items.relevant_items = [x.lower() for x in self.items.relevant_items]

        with open(RECIPES_JSON, "r") as f:
            recipes: dict[str, list[str]] = json.loads(f.read())

        # Randomly decide whether to collapse a recipe with the next one.
        self.recipes = recipes.copy()
        for result, ingredients in recipes.items():
            choosing = ingredients
            # Remove any ingredients that cannot be crafted.
            choosing = [x for x in choosing if x in self.recipes]
            # Remove any ingredients that are within our list of relevant items.
            choosing = [x for x in choosing if x not in self.items.relevant_items]

            for ingredient in choosing:
                if not self.coin_flip(0.1):
                    continue

                debug(f"Collapsing {ingredient} = {self.recipes[ingredient]}")

                # Replace the ingredient with its recipe.
                index = ingredients.index(ingredient)
                ingredients = (
                    ingredients[:index]
                    + self.recipes[ingredient]
                    + ingredients[index + 1 :]
                )

                debug(f"  new ingredients: {ingredients}")
                self.recipes[result] = ingredients
                del self.recipes[ingredient]

        self.wanted = self.rand.sample(self.items.relevant_items, 6)

    def generate_input(self, output: typing.IO | None = None):
        print(f'wanted: {", ".join(self.wanted)}', file=output)
        print("", file=output)
        for result, ingredients in self.recipes.items():
            print(f"{result} = {' + '.join(ingredients)}", file=output)

    class DFS(BaseModel):
        ingredients: set[str] = set()
        ingredient_leaves: set[str] = set()

        def run(self, problem: "Problem", item: str, level=0):
            # Base case
            if item not in problem.recipes:
                debug(f"{'| ' * level}crafting {item} requires nothing")
                self.ingredient_leaves.add(item)
                return

            # Recursive case
            debug(
                f"{'| ' * level}crafting {item} requires {' + '.join(problem.recipes[item])}"
            )
            for ingredient in problem.recipes[item]:
                self.ingredients.add(ingredient)
                self.run(problem, ingredient, level + 1)

    def part1_answer(self) -> int:
        """
        How many items are needed to craft the first item in wanted?
        """
        dfs = self.DFS()
        dfs.run(self, self.wanted[0])
        debug(f"{self.wanted[0]} had leaves: {dfs.ingredient_leaves}")
        return len(dfs.ingredients)

    def part2_answer(self) -> int:
        """
        How many items are needed to craft all the items in wanted?
        """
        dfs = self.DFS()
        for wanted in self.wanted:
            dfs.run(self, wanted)
        debug(f"all wanted items had leaves: {dfs.ingredient_leaves}")
        return len(dfs.ingredient_leaves)


if __name__ == "__main__":
    problem_utils.main(Problem)
