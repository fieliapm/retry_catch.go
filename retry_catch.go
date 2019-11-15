// retry_catch.go - Golang implementation of simple retry framework
// Copyright (C) 2019-present Himawari Tachibana <fieliapm@gmail.com>
//
// This file is part of retry_catch.go
//
// retry_catch.go is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package retry_catch

import (
	"time"
)

type RetryFunc func() error
type CatchFunc func(attemptCount uint, err error) (bool, time.Duration)
type FinallyFunc func(r interface{})

type config struct {
	retryFunc   RetryFunc
	catchFunc   CatchFunc
	finallyFunc FinallyFunc
}

type Option func(conf *config)

func Catch(catchFunc CatchFunc) Option {
	return func(conf *config) {
		conf.catchFunc = catchFunc
	}
}

func Finally(finallyFunc FinallyFunc) Option {
	return func(conf *config) {
		conf.finallyFunc = finallyFunc
	}
}

func Try(retryFunc RetryFunc, opts ...Option) error {
	conf := config{retryFunc: retryFunc}
	for _, opt := range opts {
		opt(&conf)
	}

	defer func() {
		r := recover()

		if conf.finallyFunc != nil {
			conf.finallyFunc(r)
		}

		if r != nil {
			panic(r)
		}
	}()

	for attemptCount := uint(0); ; {
		err := conf.retryFunc()
		if err == nil {
			return nil
		}

		attemptCount++

		retryFlag := false
		var delayTime time.Duration

		if conf.catchFunc != nil {
			retryFlag, delayTime = conf.catchFunc(attemptCount, err)
		}

		if !retryFlag {
			return err
		}

		time.Sleep(delayTime)
	}
}

// preset function

func BackOffDelay(maxAttemptCount uint, delayTime time.Duration) CatchFunc {
	return func(attemptCount uint, err error) (bool, time.Duration) {
		if attemptCount < maxAttemptCount {
			return true, delayTime * time.Duration(1<<(attemptCount-1))
		} else {
			var defaultDelayTime time.Duration
			return false, defaultDelayTime
		}
	}
}
