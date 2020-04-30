// retry_catch.go - Golang implementation of simple retry framework
//
// Copyright (C) 2019-present Himawari Tachibana <fieliapm@gmail.com>
//
// This file is part of retry_catch.go
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
