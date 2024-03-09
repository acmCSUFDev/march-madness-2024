# Pairing Peers

It's time to kick off March Madness, and unfortunately 
for you, you have been given one of the hardest tasks 
that could possibly be assigned to you. That is,
getting the computer science majors to communicate 
with one another. Fortunately for you, most of the work
has been done for you! 

All of the events to get people collaborating with each
other have already been planned out by the event organizers.
Your job now is to group together students in a way that 
encourages them to meet new people and get them excited
for March Madness! To complete this task, the organizers
have given their list containing the information of
all of the people who signed up for the event.

The list will look like this:
```
CWID   Name     Email     Year     FoodPreference
----------------------------------------------------------
571936 Sam sam@csu.fullerton.edu 4 Meat
399467 Kim kim@csu.fullerton.edu 2 Vegan
174839 Justin justin@csu.fullerton.edu 6 Vegan
116398 Joel Joel@csu.fullerton.edu 2 Meat
427081 Sawyer sawyer@csu.fullerton.edu 3 Vegan
```

For our first activity, we want to form big groups of people
and have them all play [Wink Murder](https://en.wikipedia.org/wiki/Wink_murder) 
to break the ice. To keep things simple, we'll create 26 different groups here. Each group will be formed
by picking students who have the same first letter of their name.
In our sample input, we would have the following groups:
```
Group 1:
571936 Sam sam@csu.fullerton.edu 4 Meat
427081 Sawyer sawyer@csu.fullerton.edu 3 Vegan

Group 2:
174839 Justin justin@csu.fullerton.edu 6 Vegan
116398 Joel Joel@csu.fullerton.edu 2 Meat

Group 3:
399467 Kim kim@csu.fullerton.edu 2 Vegan
```
Adding together the largest CWID of each group gives us `571936 + 174839 + 399467 = 1146242`

Take a look at the list that the event organizers gave you. 
**What is the sum of the largest CWIDs for every group of students?**

## Part 2

Wink Murder was a success! Everyone had a great time,
but during the game, a problem occured. The event organizers
pulled you aside and informed you of the issue, telling you
that their plan B is to play a rock paper scissors tournament
in order to stall for time while they fix the issues. 

These pairs
of people must conform to the same rule - their names must start
with the same first letter. Grouping these people is extremely easy,
but the organizers are worried that one or even two tournaments won't 
be enough time to fix their issue.

You decide to show them a crazy statistic to calm
their nerves and assure them you were the right person
for the job. **How many distinct pairs of people can 
you group together?**

<!---
Notes for input creation:
- Every group should be an even number
- Names are 3-12 letters
- Emails are name@csu.fullerton.edu
- Food Preferences are either Vegan or Meat
- Year is 1-6
- No CWID should be repeated
- Every group should have min 8 people
-->
