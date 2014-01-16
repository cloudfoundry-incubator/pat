package redis_test

import (
	"fmt"
	. "github.com/julz/pat/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Redis", func() {
	Describe("Stub", func() {
		It("Sends messages to time a function remotely", func() {
			out := &DummyOutputConnection{}
			in := &DummyReplyConnection{}
			NewWorker(out, in, "a-channel-name", "a-reply-channel").Time("foo")
			Ω(out.Messages).Should(HaveLen(1))
			Ω(out.Messages[0]).Should(Equal("RPUSH a-channel-name, a-reply-channel, foo"))
		})

		It("Returns the reply time received from the reply channel", func() {
			out := &DummyOutputConnection{}
			in := &DummyReplyConnection{"BLPOP", int64(2 * time.Second)}
			time, err := NewWorker(out, in, "a-channel-name", "a-reply-channel").Time("foo")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(time.Seconds()).Should(Equal(2.0))
		})

		PIt("Times out and sends an error if a response doesn't come back quickly enough", func() {
		})
	})

	Describe("Slave", func() {
		It("Loads tasks from the queue and times them", func() {
			out := &DummyOutputConnection{}
			in := &DummyReplyConnection{"BLPOP", "a-reply-channel,foo"}
			slave := NewSlave(in, out, "a-channel-name").WithExperiment("foo", func() (time.Duration, error) {
				return time.Second * 2, nil
			})
			err := slave.Next()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(out.Messages).Should(HaveLen(1))
			Ω(out.Messages[0]).Should(Equal("RPUSH a-reply-channel, 2s"))
		})

		PIt("Sends errors back over the channel", func() {})
	})
})

type DummyOutputConnection struct {
	Messages []string
}

type DummyReplyConnection struct {
	op    string
	reply interface{}
}

func (self *DummyOutputConnection) Do(op string, args ...interface{}) (reply interface{}, err error) {
	self.Messages = append(self.Messages, op+" "+join(args))
	return nil, nil
}

func (self *DummyReplyConnection) Do(op string, args ...interface{}) (reply interface{}, err error) {
	if op == self.op {
		return self.reply, nil
	}
	return 9, nil
}

func join(strings []interface{}) (result string) {
	for i, s := range strings {
		if i > 0 {
			result = result + ", "
		}
		result = result + fmt.Sprintf("%v", s)
	}
	return result
}
