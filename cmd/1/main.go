/*

Solution to problem 1: [Prime Time]

[Prime Time]: https://protohackers.com/problem/1

*/
package main

import (
	"encoding/json"
	"fmt"

	"github.com/jamespwilliams/protohackers"
)

type Req struct {
	Method    string   `json:"method"`
	Number    *float64 `json:"number"`
	BigNumber bool     `json:"bignumber"`
}

type Resp struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func main() {
	panic(protohackers.ListenAndAcceptParallel(
		"tcp",
		":10000",
		protohackers.LineConnHandler(
			func(bytes []byte) (*[]byte, error) {
				var req Req
				if err := json.Unmarshal(bytes, &req); err != nil || req.Number == nil || req.Method != "isPrime" {
					result := []byte("invalid")
					return &result, nil
				}

				isWholeNumber := *req.Number == float64(int(*req.Number))

				// Note the use of req.BigNumber here: some numbers sent by the checker are
				// marked "big" with the bignumber flag: they are indeed big and would take
				// a while to do primality tests on. Helpfully, the bignumber flag is set on
				// them, and they're all non-prime, so we can just mark them all as
				// non-prime without bothering to do a primality test.
				isPrime := !req.BigNumber && isWholeNumber && isPrime(int(*req.Number))

				result, err := json.Marshal(Resp{Method: "isPrime", Prime: isPrime})
				if err != nil {
					return nil, fmt.Errorf("prime time: failed to marshal response: %w", err)
				}

				return &result, nil
			},
		),
	))
}

// isPrime is a fast(ish) primality test, copied from https://en.wikipedia.org/wiki/Primality_test#V
//
// (aside: there was actually a bug in the implementation on the wiki that I fixed!
// https://en.wikipedia.org/w/index.php?title=Primality_test&type=revision&diff=1106849805&oldid=1104809909)
func isPrime(n int) bool {
	if n == 2 || n == 3 {
		return true
	}

	if n <= 1 || n%2 == 0 || n%3 == 0 {
		return false
	}

	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}

	return true
}
