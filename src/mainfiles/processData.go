package mainfiles

import (
	"time"
)

func processData() {
	switch to.internalQueue {
	case true:
		if pause {
			time.Sleep(1 * time.Second)
		}
		s := <-to.queue
		postData(s)
	case false:
		s := <-to.queue

		if len(to.queue) > 5000 {
			lobi().Send([]byte(s))
		} else {
			if pause {
				time.Sleep(1 * time.Second)
			}
			postData(s)
		}
	}
}
