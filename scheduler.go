
package main

import (
	"fmt"
	cr "github.com/robfig/cron"
	"time"
)

type Scheduler struct {
	cron *cr.Cron
	ticker chan string
	subCh chan chan string
	unsubCh chan chan string
	stopCh chan struct{}
	subscribers map[chan string] struct{}
}

func NewScheduler() *Scheduler {
	ticker := make(chan  string, 1)
	subCh := make(chan chan string, 1)
	unsubCh := make(chan chan string, 1)
	stopCh := make(chan  struct{})
	subscribers :=  make(map[chan string] struct{})
	cron := cr.New()
	cron.Start();
	cron.AddFunc("@every 1s", func() {
		select{
		case ticker <- "1s":
		default:
		}
	})
	cron.AddFunc("@every 3s", func() {
		select{
		case ticker <- "3s":
		default:
		}
	})
	go func(){
		for {
			select{
				case <- stopCh:
					for subscriber := range subscribers {
						close(subscriber)
					}
					return
				case subscriber := <- subCh:
					subscribers[subscriber] = struct{}{}
				case unsubscriber := <- unsubCh:
					delete(subscribers,unsubscriber)
					close(unsubscriber)
				case msg := <- ticker:
					for subscriber := range subscribers{
						select{
						  	case subscriber <- msg:
							default:
						}
					}
			}
		}
	}()
	return &Scheduler{
		cron,
		ticker,
		subCh,
		unsubCh,
		stopCh,
		subscribers,
	}
}

func (s *Scheduler) Subscribe() chan string  {
	subscriber := make(chan string, 5)
	s.subCh <- subscriber
	return subscriber;
}

func (s *Scheduler) Unsubscribe(c chan string) {
	s.unsubCh <- c
}

func (s *Scheduler) Start()  {
	if s.cron != nil {
		s.cron.Start()
	}
}

func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

func (s *Scheduler) Close() {
	s.Stop()
	close(s.stopCh)
}

func main() {
 	s := NewScheduler()

 	// closes scheduler in 10 seconds
 	go func(){
		time.Sleep(10 * time.Second)
		s.Close()
		fmt.Println("Close scheduler")
	}()

	// subscribes another listener in separated routine and unsubscribe in 5 seconds
	go func(){
		subs2 := s.Subscribe()
		fmt.Println("Sibscribe secondary listener")
		go func(){
			time.Sleep(5 * time.Second)
			s.Unsubscribe(subs2)
			fmt.Println("Unsibscribe secondary listener")
		}()
		for tick := range subs2{
			fmt.Println("Secondary: ", tick)
		}
	}()

	// subscribes listener
 	subs1 := s.Subscribe()
	fmt.Println("Sibscribe main listener")
 	for tick := range subs1{
		fmt.Println("Main: ",tick)
	}

	fmt.Println("Done")
}
