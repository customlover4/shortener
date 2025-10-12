package service

import (
	"math/rand/v2"
	"strings"
	"time"
)

var (
	seed = rand.NewPCG(
		uint64(time.Now().UnixMicro()),
		uint64(time.Now().Add(time.Millisecond*100).UnixMicro()),
	)
)

func generateAlias(l int) string {
	res := strings.Builder{}
	res.Grow(l)
	rd := rand.New(
		seed,
	)

	alphabet := "qwertyuiopasdfghjklzxcvbnm1234567890"
	for i := 0; i < l; i++ {
		res.WriteByte(alphabet[rd.IntN(len(alphabet))])
	}
	return res.String()
}
