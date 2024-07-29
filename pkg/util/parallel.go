package util

// ParBatchJob represents an atomic division of work which is composed of one or
// more jobs.  The idea is that all of these jobs must be computed together in
// one large batch, and cannot be further broken down.
type ParBatchJob interface {
	// Get the job identifies for all jobs in this batch.
	Jobs() []uint
	// Get the jobs on which this at least one job in this batch dependend.  In
	// otherwords, all of the return jobs must be complete before this batch can
	// run.
	Dependencies() []uint
	// Run this batch job
	Run() error
}

// ParExec executes a set of jobs in parallel using go-routines.
func ParExec[J ParBatchJob](worklist []J) error {
	var next J
	// Initialise the done set
	todo := initToDoList(worklist)
	// Iterate until all batches complete
	for len(worklist) > 0 {
		next, worklist = selectBatch(todo, worklist)
		// Execute next batch
		if err := next.Run(); err != nil {
			return err
		}
		// Mark all jobs in batch as done
		for _, j := range next.Jobs() {
			todo[j] = false
		}
	}
	// Done
	return nil
}

// Initialise the set of jobs which remain to be completed.  Jobs which are not
// present in the batch are assumed to be already completed.
func initToDoList[J ParBatchJob](batches []J) []bool {
	n := uint(0)
	// Determine largest job identifier
	for _, b := range batches {
		for _, j := range b.Jobs() {
			n = max(n, j+1)
		}
	}
	// Construct todo list
	todo := make([]bool, n)
	// Initialise jobs
	for _, b := range batches {
		for _, j := range b.Jobs() {
			todo[j] = true
		}
	}
	// Done
	return todo
}

// Select and remove the first "ready" job from the worklist.  A job is ready if
// all of its dependencies are completed.  If no such job exists, then its an
// error.
func selectBatch[J ParBatchJob](todo []bool, worklist []J) (J, []J) {
	for i, b := range worklist {
		if readyJob(todo, b) {
			// Following mechanism for removing element from worklist is
			// temporary.  It works on the assumption of sequential execution,
			// where the head is always selected first.  This obviously won't
			// work for true parallel execution.
			if i != 0 {
				panic("internal failure")
			}
			//
			return b, worklist[1:]
		}
	}
	//
	panic("no job is ready to run")
}

// ReadyJob determines whether or not a given batch job is ready to run, or not.
// Specifically, a job is ready when all its dependencies have been completed.
func readyJob[J ParBatchJob](todo []bool, batch J) bool {
	// Check dependencies
	for _, j := range batch.Dependencies() {
		if todo[j] {
			// Dependent job remains to be done.  Therefore, this job is not
			// ready to run.
			return false
		}
	}
	// All dependencies done, so this batch is ready.
	return true
}
