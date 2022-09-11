/*

Solution to problem 2: [Means to an End]

[Means to an End]: https://protohackers.com/problem/2

*/
package main

import (
	"encoding/binary"
	"math/big"

	"github.com/jamespwilliams/protohackers"
)

func main() {
	panic(protohackers.ListenAcceptAndHandleParallelStateful(
		"tcp",
		":10000",
		func() map[int32]int32 {
			return make(map[int32]int32)
		},
		protohackers.NBytesConnHandlerStateful(9, handler),
	))
}

func handler(prices map[int32]int32, bytes []byte) (*[]byte, error) {
	if len(bytes) < 9 {
		return nil, nil
	}

	t := bytes[0]

	arg1 := int32(big.NewInt(0).SetBytes(bytes[1:5]).Int64())
	arg2 := int32(big.NewInt(0).SetBytes(bytes[5:]).Int64())

	switch t {
	case 'I':
		prices[arg1] = arg2
	case 'Q':
		sum := int64(0)
		count := int64(0)
		for time, price := range prices {
			if time >= arg1 && time <= arg2 {
				sum += int64(price)
				count += 1
			}
		}

		var avg int32
		if count != 0 {
			avg = int32(sum / count)
		}

		var b [4]byte
		binary.BigEndian.PutUint32(b[:], uint32(avg))
		r := b[:]
		return &r, nil
	}

	return nil, nil
}
