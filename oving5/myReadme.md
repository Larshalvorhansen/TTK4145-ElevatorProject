## Reflecting 

You do not have to answer every question in turn, as long as you address the contents somewhere.

- Condition variables, Java monitors, and Ada protected objects are quite similar in what they do (temporarily yield execution so some other task can unblock us).  
  - But in what ways do these mechanisms differ?

    They all kinda do the same thing, but Java is built-in and easy, Ada is strict and safe, and condition variables are more manual and easy to mess up.

- Bugs in this kind of low-level synchronization can be hard to spot.  
  - Which solutions are you most confident are correct?  
  - Why, and what does this say about code quality?

    The Go solution with channels and a priority queue felt most solid. It's clear who waits and when. That probably means the code is decent.

- We operated only with two priority levels here, but it makes sense for this "kind" of priority resource to support more priorities.  
  - How would you extend these solutions to N priorities? Is it even possible to do this elegantly?  
  - What (if anything) does that say about code quality?

    With the queue, N priorities just work. Sorting by number is enough. So yeah, if the design is good, more priorities aren't a big deal.

- In D's standard library, `getValue` for semaphores is not even exposed (probably because it is not portable – Windows semaphores don't have `getValue`, though you could hack it together with `ReleaseSemaphore()` and `WaitForSingleObject()`).  
  - A leading question: Is using `getValue` ever appropriate?  
  - Explain your intuition: What is it that makes `getValue` so dubious?

    `getValue` feels sketchy. It can say "1", but the resource might be gone a moment later. It gives a false sense of control. Better to block and wait.

- Which one(s) of these different mechanisms do you prefer, both for this specific task and in general?

    For this task, I liked Go channels. In general, I just go with whatever is easiest to not mess up.
