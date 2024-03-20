# Internet Connection

March madness has started and everything is going perfectly as planned! Well, minus
some issues with the coding questions... Because you're such an amazing
organizer and the lovely people at ACM offer so much support, the event is far
more popular than anyone could have imagined.

It's nearly time to host the Hackathon, and you're beginning to realize that
it's almost impossible to pack everyone into the Computer Science building for
this event. The concern now becomes, where can we put students on campus so
that they still have access to a strong Internet connection throughout the
Hackathon?

Thankfully, you managed to get a list of all the locations of available routers
on campus, as well as how far they can reach. The list of all routers may look
something like the following:

```
Router located at x=4, y=7 with reach=3ft
Router located at x=6, y=4 with reach=4ft
Router located at x=10, y=10 with reach=3ft
```

Imagining the campus as a grid on a Cartesian (2D) plane, we can picture where
the routers are. Unfortunately for us, the school cheaped out when buying
routers, so they opted for a large quantity of routers with bad service, but
long range.

A single router is not enough for a student to maintain a stable connection. To
combat this while remaining cost-effective, the school has implemented a system
where two routers can help boost each other's signals and offer a student a
stable connection!

If a student is within range of two different routers at once, their connection
will be more than enough to sustain any activities they may try. Our goal is to
ensure that every student has a stable connection for this Hackathon, so we
will only look for areas that are in the range of 2 different routers to try
and place them there.

## Part 1

Given the long list of routers, **how many pairs of routers offer an area of
any size where a student can maintain a stable connection?**

## Part 2

To avoid scattering students all over campus in search of a good Internet
connection, we have to optimize our space. We have all the spaces where
students have a good Internet connection now, and we just need to figure out
where the students should go. **What is the largest area of overlap that any
pair of routers creates?**
