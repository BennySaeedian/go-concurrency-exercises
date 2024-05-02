//////////////////////////////////////////////////////////////////////
//
// Given is a producer-consumer scenario, where a producer reads in
// tweets from a mockstream and a consumer is processing the
// data. Your task is to change the code so that the producer as well
// as the consumer can run concurrently
//

package main

import (
	"fmt"
	"time"
)

func producer(stream Stream, tweets chan<- *Tweet) {
	// Consume the stream until exhaustion, gathering tweets
	// close the stream upon the EOF
	for {
		tweet, err := stream.Next()
		if err == ErrEOF {
			close(tweets)
			return
		}
		tweets <- tweet
	}
}

func consumer(tweets <-chan *Tweet, done chan<- bool) {
	for tweet := range tweets {
		if tweet.IsTalkingAboutGo() {
			fmt.Println(tweet.Username, "\ttweets about golang")
		} else {
			fmt.Println(tweet.Username, "\tdoes not tweet about golang")
		}
	}
	done <- true
}

func main() {
	start := time.Now()
	stream := GetMockStream()
	done := make(chan bool)
	tweets := make(chan *Tweet)

	// Run the producer concurrently
	go producer(stream, tweets)

	// Run the consumer concurrently as well
	go consumer(tweets, done)
	// wait for consumer to finish processing
	<-done

	fmt.Printf("Process took %s\n", time.Since(start))
}
