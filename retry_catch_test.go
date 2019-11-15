// retry_catch_test.go - test program for simple retry framework
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

package retry_catch_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/fieliapm/retry_catch.go"
)

var (
	testError    = errors.New("test error")
	backOffDelay retry_catch.CatchFunc
)

func init() {
	backOffDelay = retry_catch.BackOffDelay(3, time.Second)
}

func TestDefaultSuccess(t *testing.T) {
	var tryExecutionCount uint = 0

	err := retry_catch.Try(
		func() error {
			tryExecutionCount++
			if tryExecutionCount < 1 {
				return testError
			} else {
				return nil
			}
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), tryExecutionCount)
}

func TestDefaultFail(t *testing.T) {
	var tryExecutionCount uint = 0

	err := retry_catch.Try(
		func() error {
			tryExecutionCount++
			if tryExecutionCount < 2 {
				return testError
			} else {
				return nil
			}
		},
	)

	assert.Equal(t, testError, err)
	assert.Equal(t, uint(1), tryExecutionCount)
}

func TestDefaultCatchSuccess(t *testing.T) {
	var finallyExecutionCount uint = 0
	defer func() {
		if finallyExecutionCount != 1 {
			assert.Fail(t, "finally is not executed exactly 1 time")
		}
	}()

	var tryExecutionCount uint = 0

	err := retry_catch.Try(
		func() error {
			tryExecutionCount++
			if tryExecutionCount < 1 {
				return testError
			} else {
				return nil
			}
		},
		retry_catch.Finally(func(r interface{}) {
			finallyExecutionCount++
			assert.Equal(t, nil, r)
		}),
	)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), tryExecutionCount)
}

func TestDefaultCatchFail(t *testing.T) {
	var finallyExecutionCount uint = 0
	defer func() {
		if finallyExecutionCount != 1 {
			assert.Fail(t, "finally is not executed exactly 1 time")
		}
	}()

	var tryExecutionCount uint = 0

	err := retry_catch.Try(
		func() error {
			tryExecutionCount++
			if tryExecutionCount < 2 {
				return testError
			} else {
				return nil
			}
		},
		retry_catch.Finally(func(r interface{}) {
			finallyExecutionCount++
			assert.Equal(t, nil, r)
		}),
	)

	assert.Equal(t, testError, err)
	assert.Equal(t, uint(1), tryExecutionCount)
}

func TestRetry3TimesSuccessAt1stTime(t *testing.T) {
	var tryExecutionCount uint = 0

	err := retry_catch.Try(
		func() error {
			tryExecutionCount++
			if tryExecutionCount < 1 {
				return testError
			} else {
				return nil
			}
		},
		retry_catch.Catch(func(attemptCount uint, err error) (bool, time.Duration) {
			assert.Fail(t, "should not catch any error")

			return backOffDelay(attemptCount, err)
		}),
	)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), tryExecutionCount)
}

func TestRetry3TimesSuccessAt3rdTime(t *testing.T) {
	var tryExecutionCount uint = 0

	err := retry_catch.Try(
		func() error {
			tryExecutionCount++
			if tryExecutionCount < 3 {
				return testError
			} else {
				return nil
			}
		},
		retry_catch.Catch(func(attemptCount uint, err error) (bool, time.Duration) {
			assert.Equal(t, tryExecutionCount, attemptCount)
			assert.Equal(t, testError, err)

			return backOffDelay(attemptCount, err)
		}),
	)

	assert.NoError(t, err)
	assert.Equal(t, uint(3), tryExecutionCount)
}

func TestRetry3TimesFail(t *testing.T) {
	var tryExecutionCount uint = 0

	err := retry_catch.Try(
		func() error {
			tryExecutionCount++
			if tryExecutionCount < 4 {
				return testError
			} else {
				return nil
			}
		},
		retry_catch.Catch(func(attemptCount uint, err error) (bool, time.Duration) {
			assert.Equal(t, tryExecutionCount, attemptCount)
			assert.Equal(t, testError, err)

			return backOffDelay(attemptCount, err)
		}),
	)

	assert.Equal(t, testError, err)
	assert.Equal(t, uint(3), tryExecutionCount)
}

func TestRetry3TimesPanic(t *testing.T) {
	var finallyExecutionCount uint = 0
	defer func() {
		if finallyExecutionCount != 1 {
			assert.Fail(t, "finally is not executed exactly 1 time")
		}
	}()

	var tryExecutionCount uint = 0

	assert.Panics(t,
		func() {
			_ = retry_catch.Try(
				func() error {
					tryExecutionCount++
					if tryExecutionCount < 4 {
						panic(testError)
					} else {
						return nil
					}
				},
				retry_catch.Catch(func(attemptCount uint, err error) (bool, time.Duration) {
					assert.Equal(t, tryExecutionCount, attemptCount)
					assert.Equal(t, testError, err)

					return backOffDelay(attemptCount, err)
				}),
				retry_catch.Finally(func(r interface{}) {
					finallyExecutionCount++
					assert.Equal(t, testError, r)
				}),
			)

			assert.Fail(t, "should not execute code after Try() while panic")
		},
	)

	assert.Equal(t, uint(1), tryExecutionCount)
}
