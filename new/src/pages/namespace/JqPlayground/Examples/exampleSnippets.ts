const snippets = [
  {
    key: "unchangedInput",
    example: ".",
    query: ".",
    input: '{ "foo": { "bar": { "baz": 123 } } }',
  },
  {
    key: "valueAtKey",
    example: ".foo, .foo.bar, .foo?",
    query: ".foo",
    input: '{"foo": 42, "bar": "less interesting data"}',
  },
  {
    key: "arrayOperation",
    example: ".[], .[]?, .[2], .[10:15]",
    query: ".foo[1]",
    input:
      '{"foo": [{"name":"JSON", "good":true}, {"name":"XML", "good":false}]}',
  },
  {
    key: "arrayObjectConstruction",
    example: "[], {}",
    query: "[{user, title: .titles[]}]",
    input: '{"user":"stedolan","titles":["JQ Primer", "More JQ"]}',
  },
  {
    key: "lengthOfValue",
    example: "length",
    query: "[.foo[] | length]",
    input: '{"foo": [[1,2], "string", {"a":2}, null]}',
  },
  {
    key: "keysInArray",
    example: "keys",
    query: "keys",
    input: '{"abc": 1, "abcd": 2, "Foo": 3}',
  },
  {
    key: "feedInput",
    example: ",",
    query: "[.foo, .bar]",
    input: '{ "foo": 42, "bar": "something else", "baz": true}',
  },
  {
    key: "pipeOutput",
    example: "|",
    query: "[.foo[] | .name]",
    input:
      '{"foo": [{"name":"JSON", "good":true}, {"name":"XML", "good":false}]}',
  },
  {
    key: "inputUnchanged",
    example: "select(foo)",
    query: "map(select(. >= 2))",
    input: '{"a": 1, "b": 2, "c": 4, "d": 7}',
  },
  {
    key: "invokeFilter",
    example: "map(foo)",
    query: "map(.+1)",
    input: '{"a": 1, "b": 2, "c": 3}',
  },
  {
    key: "conditionals",
    example: "if-then-else-end",
    query: 'if .foo == 0 then "zero" elif .foo == 1 then "one" else "many" end',
    input: '{"foo": 2}',
  },
  {
    key: "stringInterpolation",
    example: "(foo)",
    query: '"The input was \\(.input), which is one less than \\(.input+1)"',
    input: '{"input": 42}',
  },
] as const;
export default snippets;
