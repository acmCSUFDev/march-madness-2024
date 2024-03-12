#!/usr/bin/env -S deno run -A
/// <reference lib="deno.ns" />

import type { Recipe, WorkerMessage } from "./generate_recipes_worker.ts";
import * as path from "https://deno.land/std@0.219.0/path/mod.ts";
import { sha256 } from "https://deno.land/x/sha256@v1.0.2/mod.ts";
import items from "./items.json" with { type: "json" };

const outputRecipesFile = path.join(baseDir(), "recipes.json");

async function main() {
  const wantedRecipes = await Promise.all(
    [...items.relevant_items, ...items.miscellaneous_items].map((item) =>
      findRecipesForItem(item).catch((err) => {
        throw new Error(`Failed to find item ${item}: ${err}`);
      }),
    ),
  );

  const outputRecipes: { [key: string]: string[] } = {};
  for (const recipes of wantedRecipes) {
    for (const recipe of recipes) {
      if (recipe.result in outputRecipes) {
        const existing = outputRecipes[recipe.result];
        if (!arrayEquals(existing, [recipe.first, recipe.second])) {
          throw new Error("Recipe conflict: " + recipe.result);
        }
      }
      // outputRecipes.push([recipe.result, recipe.first, recipe.second]);
      outputRecipes[recipe.result] = [recipe.first, recipe.second];
    }
  }

  console.log("Got a total of", Object.entries(outputRecipes).length, "recipes");

  // Deduplicate all recipe ingredients.
  for (const [result, ingredients] of Object.entries(outputRecipes)) {
    outputRecipes[result] = [...new Set(ingredients)];
  }

  // Get the custom JSON formatting that we want.
  const lines = [];
  for (const recipe of Object.entries(outputRecipes)) {
    const [result, ...ingredients] = [recipe[0], ...recipe[1]]
      .map((s) => sanitizeName(s))
      .map((s) => JSON.stringify(s));
    lines.push(`  ${result}: [${ingredients.join(", ")}],`);
  }
  lines.sort();
  lines.push(lines.pop().slice(0, -1)); // remove trailing comma
  const output = "{\n" + lines.join("\n") + "\n}\n";

  await Deno.writeTextFile(outputRecipesFile, output);
  console.log("Hash of output:", sha256(output, undefined, "base64"));
}

function sanitizeName(name: string): string {
  return name.toLowerCase().replaceAll("+", "p").replaceAll("=", "?");
}

function arrayEquals<T>(a: T[], b: T[]): boolean {
  if (a.length !== b.length) {
    return false;
  }
  a = a.slice().sort();
  b = b.slice().sort();
  return a.every((v, i) => v === b[i]);
}

async function findRecipesForItem(item: string): Promise<Recipe[]> {
  const worker = new Worker(import.meta.resolve("./generate_recipes_worker.ts"), {
    type: "module",
  });
  const promise = new Promise<Recipe[]>((resolve, reject) => {
    worker.onmessage = (event: MessageEvent<WorkerMessage>) => {
      if (event.data.ok === true) {
        resolve(event.data.recipes);
        return;
      }
      if (event.data.ok === false) {
        reject(event.data.error);
        return;
      }
    };
  });
  worker.postMessage({ item });
  try {
    const recipes = await promise;
    console.log("Got recipes for", item, "containing", recipes.length, "recipes");
    return recipes;
  } finally {
    worker.terminate();
  }
}

// baseDir returns the directory of the current file.
function baseDir() {
  return new URL(".", import.meta.url).pathname;
}

await main();
