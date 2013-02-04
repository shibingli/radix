package redis

import (
	. "launchpad.net/gocheck"
	"bytes"
	"bufio"
)

type ParserSuite struct{}
var _ = Suite(&ParserSuite{})

func parseString(s string) *Reply {
	reader := bufio.NewReader(bytes.NewBufferString(s))
	return parse(reader)
}

func (s *ParserSuite) TestParse(c *C) {
	// missing \n trailing
	r := parseString("foo")
	c.Check(r.Type, Equals, ErrorReply)
	c.Check(r.Err, NotNil)

	// error reply
	r = parseString("-ERR unknown command 'foobar'\r\n")
	c.Check(r.Type, Equals, ErrorReply)
	c.Check(r.Err.Error(), Equals, "ERR unknown command 'foobar'")

	// LOADING error
	r = parseString("-LOADING Redis is loading the dataset in memory\r\n")
	c.Check(r.Type, Equals, ErrorReply)
	c.Check(r.Err, Equals, LoadingError)

	// status reply
	r = parseString("+OK\r\n")
	c.Check(r.Type, Equals, StatusReply)
	c.Check(r.str, Equals, "OK")

	// integer reply
	r = parseString(":1337\r\n")
	c.Check(r.Type, Equals, IntegerReply)
	c.Check(r.int, Equals, int64(1337))

	// null bulk reply
	r = parseString("$-1\r\n")
	c.Check(r.Type, Equals, NilReply)

	// bulk reply
	r = parseString("$6\r\nfoobar\r\n")
	c.Check(r.Type, Equals, BulkReply)
	c.Check(r.str, Equals, "foobar")

	// null multi bulk reply
	r = parseString("*-1\r\n")
	c.Check(r.Type, Equals, NilReply)

	// multi bulk reply
	r = parseString("*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$6\r\nfoobar\r\n")
	c.Check(r.Type, Equals, MultiReply)
	c.Assert(len(r.Elems), Equals, 5)
	c.Check(r.Elems[0].int, Equals, int64(1))
	c.Check(r.Elems[1].int, Equals, int64(2))
	c.Check(r.Elems[2].int, Equals, int64(3))
	c.Check(r.Elems[3].int, Equals, int64(4))
	c.Check(r.Elems[4].str, Equals, "foobar")

	// invalid multi bulk reply
	r = parseString("*-2\r\n")
	c.Check(r.Type, Equals, ErrorReply)
	c.Check(r.Err, Equals, ParseError)

	// invalid reply
	r = parseString("@foo\r\n")
	c.Check(r.Type, Equals, ErrorReply)
	c.Check(r.Err, Equals, ParseError)

}
