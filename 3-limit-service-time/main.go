//////////////////////////////////////////////////////////////////////
//
// Your video processing service has a freemium model. Everyone has 10
// sec of free processing time on your service. After that, the
// service will kill your process, unless you are a paid premium user.
//
// Beginner Level: 10s max per request
// Advanced Level: 10s max per user (accumulated)
//

package main

import (
	"sync/atomic"
	"time"
)

// Free processing time allowed per user, in seconds
const (
	FreeQuantaAllowed int64         = 10
	Quantom           time.Duration = time.Second
)

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	ID        int
	IsPremium bool
	TimeUsed  int64 // in seconds
}

// handles premium user's requests
func handlePremiumUser(process func()) bool {
	// no time limit for premium users
	process()
	return true
}

// handles un-paying users trail mode requests
func handleFreeUser(process func(), u *User) bool {
	// if the user exhaused the free trial, return false
	if atomic.LoadInt64(&u.TimeUsed) >= FreeQuantaAllowed {
		return false
	}
	ticker := time.Tick(Quantom)
	processDone := make(chan bool)
	// run process in a go-routine and report once it's done
	// notice the process can be terminated before full execution
	go func() {
		process()
		processDone <- true
	}()
	// update TimeUsed every quantom, if the user exhausts the free trial, return false
	// if the process is complete before that, return true
	for {
		select {
		// user process finished before timeout
		case <-ticker:
			// atomic u.TimeUsed++
			timeUsed := atomic.AddInt64(&u.TimeUsed, 1)
			if timeUsed >= FreeQuantaAllowed {
				return false
			}
		case <-processDone:
			return true
		}
	}

}

// HandleRequest runs the processes requested by users. Returns false
// if process had to be killed
func HandleRequest(process func(), u *User) bool {
	if u.IsPremium {
		return handlePremiumUser(process)
	}
	// else, free trial model
	return handleFreeUser(process, u)
}

func main() {
	RunMockServer()
}
