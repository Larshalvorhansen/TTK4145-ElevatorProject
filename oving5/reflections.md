## Reflecting

- Condition variables, Java monitors, and Ada protected objects are quite similar in what they do (temporarily yieldexecution so some other task can unblock us). But in what ways do these mechanisms differ?

    They all kinda do the same thing, but Java is more built-in to the language "wait()/notify()", Ada is strict meaning youre code has higher requirements before it compiles "pthread_mutex_lock(&mutex)/pthread_mutex_unlock(&mutex)
" , and condition variables are more manual and easy to mess up.

- Bugs in this kind of low-level synchronization can be hard to spot.  
  - Which solutions are you most confident are correct?  

    The Go solution with channels and a priority queue felt most solid and familiar. Probably since our elevator project is written in go. It's clear who waits and when. The channels in go are great for updating variables concurrently accross the board. for example when door is obstructed: 
    "if obstruction {obstructedCh <- true}

- We operated only with two priority levels here, but it makes sense for this "kind" of priority resource to support more priorities.  
  - How would you extend these solutions to N priorities? Is it even possible to do this elegantly?  
  - What (if anything) does that say about code quality?

    With the queue, N priorities just work. Sorting by number is enough. If the design is good, more priorities aren't a big deal.
```go
func main() {
	taskCh := make(chan *Task)
	doneCh := make(chan string)

	// Start the worker
	go worker(taskCh, doneCh)

	// Create and populate the priority queue
	pq := &PriorityQueue{}
	heap.Init(pq)

	heap.Push(pq, &Task{"Make cofee", 3})
	heap.Push(pq, &Task{"Write exam", 0})
	heap.Push(pq, &Task{"Eat breakfast", 1})

	// Dispatch tasks in priority order
	for pq.Len() > 0 {
		task := heap.Pop(pq).(*Task)
		taskCh <- task
	}

	// Wait for all tasks to be completed
	for i := 0; i < 5; i++ {
		fmt.Println("Done:", <-doneCh)
	}

	close(taskCh)
	close(doneCh)
}```

- In D's standard library, `getValue` for semaphores is not even exposed (probably because it is not portable – Windows semaphores don't have `getValue`, though you could hack it together with `ReleaseSemaphore()` and `WaitForSingleObject()`).  
  - A leading question: Is using `getValue` ever appropriate?  
  - Explain your intuition: What is it that makes `getValue` so dubious?

    `getValue` feels sketchy. It can say "1", but the resource might be gone a moment later. It gives a false sense of control. Better to block and wait.

- Which one(s) of these different mechanisms do you prefer, both for this specific task and in general?

    For this task, I/we prefer using Go channels(so supprise), mostly because it is what is most familiar. 
