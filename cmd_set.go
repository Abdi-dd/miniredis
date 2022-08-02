// Commands from https://redis.io/commands#set

package miniredis

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alicebob/miniredis/v2/server"
)

// commandsSet handles all set value operations.
func commandsSet(m *Miniredis) {
	m.srv.Register("SADD", m.cmdSadd)
	m.srv.Register("SCARD", m.cmdScard)
	m.srv.Register("SDIFF", m.cmdSdiff)
	m.srv.Register("SDIFFSTORE", m.cmdSdiffstore)
	m.srv.Register("SINTER", m.cmdSinter)
	m.srv.Register("SINTERSTORE", m.cmdSinterstore)
	m.srv.Register("SISMEMBER", m.cmdSismember)
	m.srv.Register("SMEMBERS", m.cmdSmembers)
	m.srv.Register("SMOVE", m.cmdSmove)
	m.srv.Register("SPOP", m.cmdSpop)
	m.srv.Register("SRANDMEMBER", m.cmdSrandmember)
	m.srv.Register("SREM", m.cmdSrem)
	m.srv.Register("SUNION", m.cmdSunion)
	m.srv.Register("SUNIONSTORE", m.cmdSunionstore)
	m.srv.Register("SSCAN", m.cmdSscan)
}

// SADD
func (m *Miniredis) cmdSadd(c *server.Peer, cmd string, args []string) {
	if len(args) < 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	key, elems := args[0], args[1:]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if db.exists(key) && db.t(key) != "set" {
			c.WriteError(ErrWrongType.Error())
			return
		}

		added := db.setAdd(key, elems...)
		c.WriteInt(added)
	})
}

// SCARD
func (m *Miniredis) cmdScard(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	key := args[0]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(key) {
			c.WriteInt(0)
			return
		}

		if db.t(key) != "set" {
			c.WriteError(ErrWrongType.Error())
			return
		}

		members := db.setMembers(key)
		c.WriteInt(len(members))
	})
}

// SDIFF
func (m *Miniredis) cmdSdiff(c *server.Peer, cmd string, args []string) {
	if len(args) < 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	keys := args

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		set, err := db.setDiff(keys)
		if err != nil {
			c.WriteError(err.Error())
			return
		}

		c.WriteSetLen(len(set))
		for k := range set {
			c.WriteBulk(k)
		}
	})
}

// SDIFFSTORE
func (m *Miniredis) cmdSdiffstore(c *server.Peer, cmd string, args []string) {
	if len(args) < 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	dest, keys := args[0], args[1:]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		set, err := db.setDiff(keys)
		if err != nil {
			c.WriteError(err.Error())
			return
		}

		db.del(dest, true)
		db.setSet(dest, set)
		c.WriteInt(len(set))
	})
}

// SINTER
func (m *Miniredis) cmdSinter(c *server.Peer, cmd string, args []string) {
	if len(args) < 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	keys := args

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		set, err := db.setInter(keys)
		if err != nil {
			c.WriteError(err.Error())
			return
		}

		c.WriteLen(len(set))
		for k := range set {
			c.WriteBulk(k)
		}
	})
}

// SINTERSTORE
func (m *Miniredis) cmdSinterstore(c *server.Peer, cmd string, args []string) {
	if len(args) < 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	dest, keys := args[0], args[1:]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		set, err := db.setInter(keys)
		if err != nil {
			c.WriteError(err.Error())
			return
		}

		db.del(dest, true)
		db.setSet(dest, set)
		c.WriteInt(len(set))
	})
}

// SISMEMBER
func (m *Miniredis) cmdSismember(c *server.Peer, cmd string, args []string) {
	if len(args) != 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	key, value := args[0], args[1]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(key) {
			c.WriteInt(0)
			return
		}

		if db.t(key) != "set" {
			c.WriteError(ErrWrongType.Error())
			return
		}

		if db.setIsMember(key, value) {
			c.WriteInt(1)
			return
		}
		c.WriteInt(0)
	})
}

// SMEMBERS
func (m *Miniredis) cmdSmembers(c *server.Peer, cmd string, args []string) {
	if len(args) != 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	key := args[0]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(key) {
			c.WriteSetLen(0)
			return
		}

		if db.t(key) != "set" {
			c.WriteError(ErrWrongType.Error())
			return
		}

		members := db.setMembers(key)

		c.WriteSetLen(len(members))
		for _, elem := range members {
			c.WriteBulk(elem)
		}
	})
}

// SMOVE
func (m *Miniredis) cmdSmove(c *server.Peer, cmd string, args []string) {
	if len(args) != 3 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	src, dst, member := args[0], args[1], args[2]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(src) {
			c.WriteInt(0)
			return
		}

		if db.t(src) != "set" {
			c.WriteError(ErrWrongType.Error())
			return
		}

		if db.exists(dst) && db.t(dst) != "set" {
			c.WriteError(ErrWrongType.Error())
			return
		}

		if !db.setIsMember(src, member) {
			c.WriteInt(0)
			return
		}
		db.setRem(src, member)
		db.setAdd(dst, member)
		c.WriteInt(1)
	})
}

// SPOP
func (m *Miniredis) cmdSpop(c *server.Peer, cmd string, args []string) {
	if len(args) == 0 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	opts := struct {
		key       string
		withCount bool
		count     int
	}{
		count: 1,
	}
	opts.key, args = args[0], args[1:]

	if len(args) > 0 {
		v, err := strconv.Atoi(args[0])
		if err != nil {
			setDirty(c)
			c.WriteError(msgInvalidInt)
			return
		}
		if v < 0 {
			setDirty(c)
			c.WriteError(msgOutOfRange)
			return
		}
		opts.count = v
		opts.withCount = true
		args = args[1:]
	}
	if len(args) > 0 {
		setDirty(c)
		c.WriteError(msgInvalidInt)
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(opts.key) {
			if !opts.withCount {
				c.WriteNull()
				return
			}
			c.WriteLen(0)
			return
		}

		if db.t(opts.key) != "set" {
			c.WriteError(ErrWrongType.Error())
			return
		}

		var deleted []string
		for i := 0; i < opts.count; i++ {
			members := db.setMembers(opts.key)
			if len(members) == 0 {
				break
			}
			member := members[m.randIntn(len(members))]
			db.setRem(opts.key, member)
			deleted = append(deleted, member)
		}
		// without `count` return a single value
		if !opts.withCount {
			if len(deleted) == 0 {
				c.WriteNull()
				return
			}
			c.WriteBulk(deleted[0])
			return
		}
		// with `count` return a list
		c.WriteLen(len(deleted))
		for _, v := range deleted {
			c.WriteBulk(v)
		}
	})
}

// SRANDMEMBER
func (m *Miniredis) cmdSrandmember(c *server.Peer, cmd string, args []string) {
	if len(args) < 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if len(args) > 2 {
		setDirty(c)
		c.WriteError(msgSyntaxError)
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	key := args[0]
	count := 0
	withCount := false
	if len(args) == 2 {
		var err error
		count, err = strconv.Atoi(args[1])
		if err != nil {
			setDirty(c)
			c.WriteError(msgInvalidInt)
			return
		}
		withCount = true
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(key) {
			c.WriteNull()
			return
		}

		if db.t(key) != "set" {
			c.WriteError(ErrWrongType.Error())
			return
		}

		members := db.setMembers(key)
		if count < 0 {
			// Non-unique elements is allowed with negative count.
			c.WriteLen(-count)
			for count != 0 {
				member := members[m.randIntn(len(members))]
				c.WriteBulk(member)
				count++
			}
			return
		}

		// Must be unique elements.
		m.shuffle(members)
		if count > len(members) {
			count = len(members)
		}
		if !withCount {
			c.WriteBulk(members[0])
			return
		}
		c.WriteLen(count)
		for i := range make([]struct{}, count) {
			c.WriteBulk(members[i])
		}
	})
}

// SREM
func (m *Miniredis) cmdSrem(c *server.Peer, cmd string, args []string) {
	if len(args) < 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	key, fields := args[0], args[1:]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		if !db.exists(key) {
			c.WriteInt(0)
			return
		}

		if db.t(key) != "set" {
			c.WriteError(ErrWrongType.Error())
			return
		}

		c.WriteInt(db.setRem(key, fields...))
	})
}

// SUNION
func (m *Miniredis) cmdSunion(c *server.Peer, cmd string, args []string) {
	if len(args) < 1 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	keys := args

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		set, err := db.setUnion(keys)
		if err != nil {
			c.WriteError(err.Error())
			return
		}

		c.WriteLen(len(set))
		for k := range set {
			c.WriteBulk(k)
		}
	})
}

// SUNIONSTORE
func (m *Miniredis) cmdSunionstore(c *server.Peer, cmd string, args []string) {
	if len(args) < 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	dest, keys := args[0], args[1:]

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)

		set, err := db.setUnion(keys)
		if err != nil {
			c.WriteError(err.Error())
			return
		}

		db.del(dest, true)
		db.setSet(dest, set)
		c.WriteInt(len(set))
	})
}

// SSCAN
func (m *Miniredis) cmdSscan(c *server.Peer, cmd string, args []string) {
	if len(args) < 2 {
		setDirty(c)
		c.WriteError(errWrongNumber(cmd))
		return
	}
	if !m.handleAuth(c) {
		return
	}
	if m.checkPubsub(c, cmd) {
		return
	}

	var opts struct {
		key       string
		value     int
		cursor    int
		count     int
		withMatch bool
		match     string
	}

	opts.key = args[0]
	if ok := optIntErr(c, args[1], &opts.cursor, msgInvalidCursor); !ok {
		return
	}
	args = args[2:]

	// MATCH and COUNT options
	for len(args) > 0 {
		if strings.ToLower(args[0]) == "count" {
			if len(args) < 2 {
				setDirty(c)
				c.WriteError(msgSyntaxError)
				return
			}
			count, err := strconv.Atoi(args[1])
			if err != nil {
				setDirty(c)
				c.WriteError(msgInvalidInt)
				return
			}
			if count < 0 {
				setDirty(c)
				c.WriteError(msgInvalidInt)
				return
			}
			opts.count = count
			args = args[2:]
			continue
		}
		if strings.ToLower(args[0]) == "match" {
			if len(args) < 2 {
				setDirty(c)
				c.WriteError(msgSyntaxError)
				return
			}
			opts.withMatch = true
			opts.match = args[1]
			args = args[2:]
			continue
		}
		setDirty(c)
		c.WriteError(msgSyntaxError)
		return
	}

	withTx(m, c, func(c *server.Peer, ctx *connCtx) {
		db := m.db(ctx.selectedDB)
		// return _all_ (matched) keys every time
		if db.exists(opts.key) && db.t(opts.key) != "set" {
			c.WriteError(ErrWrongType.Error())
			return
		}
		members := db.setMembers(opts.key)
		if opts.withMatch {
			members, _ = matchKeys(members, opts.match)
		}
		low := opts.cursor
		end := low + opts.count
		// validate end is correct
		if opts.count == 0 || end > len(members) {
			end = len(members)
		}
		if opts.cursor > end {
			// invalid cursor
			c.WriteLen(2)
			c.WriteBulk("0") // no next cursor
			c.WriteLen(0)    // no elements
			return
		}
		slice := members[low:end]
		cursorValue := low + opts.count
		if cursorValue > end || opts.count == 0 {
			cursorValue = 0 // no next cursor
		}
		c.WriteLen(2)
		c.WriteBulk(fmt.Sprintf("%d", cursorValue))
		c.WriteLen(len(slice))
		for _, k := range slice {
			c.WriteBulk(k)
		}
		low = cursorValue
		end = end + opts.count

	})
}
