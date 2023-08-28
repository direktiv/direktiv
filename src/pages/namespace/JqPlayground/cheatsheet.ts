export default [
  {
    example: ".",
    tip: "unchanged input",
    query: ".",
    input: '{ "foo": { "bar": { "baz": 123 } } }',
  },
  {
    example: ".foo, .foo.bar, .foo?",
    tip: "value at key",
    query: ".foo",
    input: '{"foo": 42, "bar": "less interesting data"}',
  },
  {
    example: ".[], .[]?, .[2], .[10:15]",
    tip: "array operation",
    query: ".foo[1]",
    input:
      '{"foo": [{"name":"JSON", "good":true}, {"name":"XML", "good":false}]}',
  },
  {
    example: "[], {}",
    tip: "array/object construction",
    query: "[{user, title: .titles[]}]",
    input: '{"user":"stedolan","titles":["JQ Primer", "More JQ"]}',
  },
  {
    example: "length",
    tip: "length of a value",
    query: "[.foo[] | length]",
    input: '{"foo": [[1,2], "string", {"a":2}, null]}',
  },
  {
    example: "keys",
    tip: "keys in an array",
    query: "keys",
    input: '{"abc": 1, "abcd": 2, "Foo": 3}',
  },
  {
    example: ",",
    tip: "feed input into multiple filters",
    query: "[.foo, .bar]",
    input: '{ "foo": 42, "bar": "something else", "baz": true}',
  },
  {
    example: "|",
    tip: "pipe output of one filter to the next filter",
    query: "[.foo[] | .name]",
    input:
      '{"foo": [{"name":"JSON", "good":true}, {"name":"XML", "good":false}]}',
  },
  {
    example: "select(foo)",
    tip: "input unchanged if foo returns true",
    query: "map(select(. >= 2))",
    input: '{"a": 1, "b": 2, "c": 4, "d": 7}',
  },
  {
    example: "map(foo)",
    tip: "invoke filter foo for each input",
    query: "map(.+1)",
    input: '{"a": 1, "b": 2, "c": 3}',
  },
  {
    example: "if-then-else-end",
    tip: "conditionals",
    query: 'if .foo == 0 then "zero" elif .foo == 1 then "one" else "many" end',
    input: '{"foo": 2}',
  },
  {
    example: "(foo)",
    tip: "string interpolation",
    query: '"The input was \\(.input), which is one less than \\(.input+1)"',
    input: '{"input": 42}',
  },
] as const;
