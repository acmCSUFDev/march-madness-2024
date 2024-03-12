# Problem Generation

This directory contains problem generators, which is in charge of generating
problem inputs and their solutions. The de facto language for problem
generators is Python.

For an example of a problem generator, see [./01/problem.py](./01/problem.py).

## Specification

The server expects a problem generator to implement the following usages:

- `$PROGRAM --seed $SEED`: generate a problem input using the given seed.
- `$PROGRAM --seed $SEED --part1`: generate the part 1 solution using the given seed.
- `$PROGRAM --seed $SEED --part2`: generate the part 2 solution using the given seed.

Currently, only Python is supported as the language for problem generators.
It would be trivial to support other languages, but it is not a priority at the
moment.

## Python

Python problem generators must have a module path of `problems.NAME.problem`.
For example, if the problem name is `01`, the module path must be
`problems.01.problem`.

To test this, simply run the following command:

```sh
python -m problems.01.problem --seed 0
```
