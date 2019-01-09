
package main

import (
	"fmt"
	cr "github.com/robfig/cron"
	"time"
)

type Scheduler struct {
	cron *cr.Cron
	ticker chan interface{}
	subCh chan chan interface{}
	unsubCh chan chan interface{}
	stopCh chan struct{}
	subscribers map[chan interface{}] struct{}
}

func NewScheduler() *Scheduler {
	ticker := make(chan  interface{}, 1)
	subCh := make(chan chan interface{}, 1)
	unsubCh := make(chan chan interface{}, 1)
	stopCh := make(chan  struct{})
	subscribers :=  make(map[chan interface{}] struct{})
	cron := cr.New()
	cron.Start();
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

func (s *Scheduler) Subscribe() chan interface{}  {
	subscriber := make(chan interface{}, 5)
	s.subCh <- subscriber
	return subscriber;
}

func (s *Scheduler) Unsubscribe(c chan interface{}) {
	s.unsubCh <- c
}

func(s *Scheduler) AddSchedule(schedule string, message interface{}) error {
	return s.cron.AddFunc(schedule, func() {
		select{
		case s.ticker <- message:
		default:
		}
	})
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
		err := s.AddSchedule("@every 3s", 3)
		if err != nil {
			return
		}
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
 	var message string = "1s"
	err := s.AddSchedule("@every 1s", message)
	if err != nil {
		return
	}
	fmt.Println("Sibscribe main listener")
 	for tick := range subs1{
 		if res,ok := tick.(string); ok == true {
			fmt.Println("Main: ",res)
		}
	}
	fmt.Println("Done")
}
