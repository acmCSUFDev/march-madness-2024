/// <reference lib="deno.ns" />

import Finder from "https://esm.sh/v135/gh/vantezzen/infinite-craft-solver@d84ece7cb7ae998967674b9254858b8b1f2d717d/apps/web/lib/Finder.ts";

export type Recipe = {
  first: string;
  second: string;
  result: string;
};

export type WorkerMessage =
  | { ok: true; recipes: Recipe[] }
  | { ok: false; error: string };

const logError = console.error;
console.error = (...args) => {
  if (args[0].startsWith("Failed to report path")) {
    // Ignore this error. We're not running in a browser.
    return;
  }
  logError(...args);
};

const _setTimeout = self.setTimeout;
self.setTimeout = (fn, delay, ...args) => {
  // Don't even bother with the `setTimeout(code)` usage.
  if (typeof fn != "function") {
    return _setTimeout(fn, delay, ...args);
  }

  // We don't need to wait for the delay, so we can just call the function.
  if (delay == 0) {
    fn(...args);
    return;
  }

  return _setTimeout(fn, delay, ...args);
};

self.onmessage = async (event: MessageEvent<{ item: string }>) => {
  // Shut this up too.
  console.log = () => {};

  const finder = new Finder();
  try {
    const recipes = (await finder.findItem(event.data.item)) as Recipe[];
    self.postMessage({ ok: true, recipes });
  } catch (err) {
    self.postMessage({ ok: false, error: err });
  }
};
