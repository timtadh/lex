package lex

import "testing"

type testToken struct {
    name string
    attribute Attribute
}

type testAttr string

func (self *testToken) Name() string {
    return self.name
}

func (self *testToken) Attribute() Attribute {
    return self.attribute
}

func (self testAttr) String() string {
    return string(self)
}

func TestLexerMatch(t *testing.T) {
    var a Attribute = testAttr("hello")
    t.Log(a)
    var c Token = &testToken{name:"test", attribute:testAttr("hello")}
    t.Log(c)
    var b ProcessMatch = func(match []byte) (bool, Token) {
        return false, nil
    }
    t.Log(b)
    tokens := Lex([]*Pattern{
        &Pattern{"x.yz", func(match []byte) (bool, Token) {
            return true, &testToken{"xyz", testAttr(string(match))}
        }},
        &Pattern{"ab|cd", func(match []byte) (bool, Token) {
            return true, &testToken{"ab-or-cd", testAttr(string(match))}
        }},
        &Pattern{"b*", func(match []byte) (bool, Token) {
            return true, &testToken{"bstar", testAttr(string(match))}
        }}}, []byte("xqyzabbbab"))
    expected := []*testToken{
        &testToken{"xyz", testAttr("xqyz")},
        &testToken{"ab-or-cd", testAttr("ab")},
        &testToken{"bstar", testAttr("bb")},
        &testToken{"ab-or-cd", testAttr("ab")}}
    i := 0
    for token := range tokens {
        t.Log(i, "found", token)
        if expected[i].Name() != token.Name() {
            t.Fatalf("Expected %v == %v", expected[i].Name(), token.Name())
        }
        if expected[i].Attribute().String() != token.Attribute().String() {
            t.Fatalf("Expected %v == %v", expected[i].Attribute(), token.Attribute())
        }
        i++
    }
}

