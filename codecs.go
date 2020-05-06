package rsmt2d

import (
	"errors"

	"github.com/vivint/infectious"
)

// Codec type
type Codec int

// Erasure codes enum:
const (
	// RSGF8 represents Reed-Solomon Codec with an 8-bit Finite Galois Field (2^8)
	RSGF8 Codec = iota
)

// Max number of chunks each code supports in a 2D square.
var CodecsMaxChunksMap = map[Codec]int{
	RSGF8: 128 * 128,
}

var infectiousCache map[int]*infectious.FEC

func init() {
	infectiousCache = make(map[int]*infectious.FEC)
}

func encode(data [][]byte, codec Codec) ([][]byte, error) {
	switch codec {
	case RSGF8:
		result, err := encodeRSGF8(data)
		return result, err
	default:
		return nil, errors.New("invalid codec")
	}
}

func encodeRSGF8(data [][]byte) ([][]byte, error) {
	var fec *infectious.FEC
	var err error
	if value, ok := infectiousCache[len(data)]; ok {
		fec = value
	} else {
		fec, err = infectious.NewFEC(len(data), len(data)*2)
		if err != nil {
			return nil, err
		}

		infectiousCache[len(data)] = fec
	}

	shares := make([][]byte, len(data))
	output := func(s infectious.Share) {
		if s.Number >= len(data) {
			shareData := make([]byte, len(data[0]))
			copy(shareData, s.Data)
			shares[s.Number-len(data)] = shareData
		}
	}

	flattened := flattenChunks(data)
	err = fec.Encode(flattened, output)

	return shares, err
}

func decode(data [][]byte, codec Codec) ([][]byte, error) {
	switch codec {
	case RSGF8:
		result, err := decodeRSGF8(data)
		return result, err
	default:
		return nil, errors.New("invalid codec")
	}
}

func decodeRSGF8(data [][]byte) ([][]byte, error) {
	var fec *infectious.FEC
	var err error
	if value, ok := infectiousCache[len(data)/2]; ok {
		fec = value
	} else {
		fec, err = infectious.NewFEC(len(data)/2, len(data))
		if err != nil {
			return nil, err
		}

		infectiousCache[len(data)/2] = fec
	}

	rebuiltShares := make([][]byte, len(data)/2)
	rebuiltSharesOutput := func(s infectious.Share) {
		rebuiltShares[s.Number] = s.DeepCopy().Data
	}

	shares := []infectious.Share{}
	for j := 0; j < len(data); j++ {
		if data[j] != nil {
			shares = append(shares, infectious.Share{Number: j, Data: data[j]})
		}
	}

	err = fec.Rebuild(shares, rebuiltSharesOutput)

	return rebuiltShares, err
}