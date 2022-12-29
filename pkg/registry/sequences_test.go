package registry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSequence(t *testing.T) {
	t.Run("test sequence", func(t *testing.T) {
		asserts := assert.New(t)
		sq := newSequences(time.Second * 5)
		err := sq.check("dsadsa", uint64(time.Now().Add(-time.Second*6).UnixNano()))
		asserts.Equal(errors.New("sequence too small"), err)
		err = sq.check("dsadsa", uint64(time.Now().Add(-time.Second*3).UnixNano()))
		asserts.Nil(err)
		err = sq.check("dsadsa", uint64(time.Now().Add(-time.Second*4).UnixNano()))
		asserts.Equal(errors.New("sequence less than or equal to last seen"), err)
		time.Sleep(time.Second * 5)
		err = sq.check("dsadsa2", uint64(time.Now().Add(-time.Second*4).UnixNano()))
		asserts.Nil(err)
		sq.retire()
		asserts.Equal(1, len(sq.seqs))
	})

}
