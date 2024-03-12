# Crafting Ingredients

Daniel, our awesome VP, needs our help to get the Hackathon feast ready! We've
got a list of (potentially) food and drink items for FullyHacks, but we need to
break down the recipes into individual ingredients to figure out our shopping
list. We need your help once again!

You're given an input consisting of two parts: a very short list of items we
want, followed by a very large list of how every item is made.

The short list of items we want starts with `wanted: `, followed by a
list of items separated by commas.

The very large list of how every item is made starts with a list of items
and their ingredients. Each item is listed on a separate line, with the created
item followed by an equals sign (`=`), followed by the items that are needed to
craft it separated by plus signs (`+`).

For example, a very small input might look like this:

```
wanted: pizza, cake

pizza = dough + sauce + cheese
dough = flour + water
cake = flour + sugar + eggs + milk
```

In this input, in order to make a `pizza`, we would need `flour`, `water`,
`sauce`, and `cheese`. Our `flour` and `water` are used to make our `dough`.

Similarly, in order to make a `cake`, we would need `flour`, `sugar`, `eggs`,
and `milk`. No other ingredients are needed to make our `cake`.

## Part 1

Your first mission: **how many ingredients total do we need to whip up the
first item on our wishlist?**

In the example above, the answer would be 5, since we're counting `dough`,
`flour`, `water`, `sauce`, and `cheese`.

## Part 2

In order to save us money, we've decided to craft as many of the ingredients as
we can ourselves. For the example above, it means we would only need to buy
`flour`, `water`, `sauce`, and `cheese`, making it a total of 4 ingredients.

With this in mind, **how many ingredients do we need to craft all of the items
on our wishlist?**
