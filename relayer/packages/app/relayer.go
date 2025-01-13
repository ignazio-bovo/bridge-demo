package app

import (
	"fmt"

	"github.com/Datura-ai/relayer/packages/app/observer"
)

func StartRelayer() {

	ch := make(chan struct{})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Observer panic: %v\n", r)
			}
			ch <- struct{}{}
		}()
		observer.StartObserver()

	}()

	// Block forever to keep the relayer running
	select {}

	// go func() {
	// 	defer func() {
	// 		if r := recover(); r != nil {
	// 			tssErr = fmt.Errorf("tss panic: %v", r)
	// 		}
	// 		ch <- struct{}{}
	// 	}()
	// 	tss.StartTss()
	// 	tssDone = true
	// }()
}
