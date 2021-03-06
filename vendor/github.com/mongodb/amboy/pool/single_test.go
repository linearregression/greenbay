package pool

import (
	"fmt"
	"testing"

	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/stretchr/testify/suite"
	"github.com/tychoish/grip"
	"github.com/tychoish/grip/level"
	"golang.org/x/net/context"
)

type SingleRunnerSuite struct {
	pool  *SingleRunner
	queue *QueueTester
	suite.Suite
}

func TestSingleWorkerSuite(t *testing.T) {
	suite.Run(t, new(SingleRunnerSuite))
}

func (s *SingleRunnerSuite) SetupSuite() {
	grip.SetThreshold(level.Info)
}

func (s *SingleRunnerSuite) SetupTest() {
	s.pool = NewSingleRunner()
	s.queue = NewQueueTester(s.pool)
}

func (s *SingleRunnerSuite) TestConstructedInstanceImplementsInterface() {
	s.Implements((*amboy.Runner)(nil), s.pool)
}

func (s *SingleRunnerSuite) TestPoolErrorsOnSuccessiveStarts() {
	s.False(s.pool.Started())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.NoError(s.pool.Start(ctx))
	s.True(s.pool.Started())

	for i := 0; i < 20; i++ {
		s.NoError(s.pool.Start(ctx))
		s.True(s.pool.Started())
	}
}

func (s *SingleRunnerSuite) TestPoolStartsAndProcessesJobs() {
	const num int = 20
	var jobs []amboy.Job

	for i := 0; i < num; i++ {
		cmd := fmt.Sprintf("echo 'task=%d'", i)
		jobs = append(jobs, job.NewShellJob(cmd, ""))
	}

	s.False(s.pool.Started())
	s.False(s.queue.Started())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.NoError(s.queue.Start(ctx))

	for _, job := range jobs {
		s.NoError(s.queue.Put(job))
	}

	s.True(s.pool.Started())
	s.True(s.queue.Started())

	s.queue.Wait()

	for _, job := range jobs {
		s.True(job.Completed())
	}
}

func (s *SingleRunnerSuite) TestQueueIsMutableBeforeStartingPool() {
	s.NotNil(s.pool.queue)
	s.False(s.pool.Started())

	newQueue := NewQueueTester(s.pool)
	s.NoError(s.pool.SetQueue(newQueue))

	s.Equal(newQueue, s.pool.queue)
	s.NotEqual(s.queue, s.pool.queue)
}

func (s *SingleRunnerSuite) TestQueueIsNotMutableAfterStartingPool() {
	s.NotNil(s.pool.queue)
	s.False(s.pool.Started())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.NoError(s.pool.Start(ctx))
	s.True(s.pool.Started())

	newQueue := NewQueueTester(s.pool)
	s.Error(s.pool.SetQueue(newQueue))

	s.Equal(s.queue, s.pool.queue)
	s.NotEqual(newQueue, s.pool.queue)
}
