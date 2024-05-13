const snippets = [
  {
    key: "unchangedInput",
    example: ".",
    jx: ".",
    data: '{ "foo": { "bar": { "baz": 123 } } }',
  },
  {
    key: "valueAtKey",
    example: ".foo, .foo.bar, .foo?",
    jx: ".foo",
    data: '{"foo": 42, "bar": "less interesting data"}',
  },
  {
    key: "arrayOperation",
    example: ".[], .[]?, .[2], .[10:15]",
    jx: ".foo[1]",
    data: '{"foo": [{"name":"JSON", "good":true}, {"name":"XML", "good":false}]}',
  },
  {
    key: "arrayObjectConstruction",
    example: "[], {}",
    jx: "[{user, title: .titles[]}]",
    data: '{"user":"stedolan","titles":["JQ Primer", "More JQ"]}',
  },
  {
    key: "lengthOfValue",
    example: "length",
    jx: "[.foo[] | length]",
    data: '{"foo": [[1,2], "string", {"a":2}, null]}',
  },
  {
    key: "keysInArray",
    example: "keys",
    jx: "keys",
    data: '{"abc": 1, "abcd": 2, "Foo": 3}',
  },
  {
    key: "feedInput",
    example: ",",
    jx: "[.foo, .bar]",
    data: '{ "foo": 42, "bar": "something else", "baz": true}',
  },
  {
    key: "pipeOutput",
    example: "|",
    jx: "[.foo[] | .name]",
    data: '{"foo": [{"name":"JSON", "good":true}, {"name":"XML", "good":false}]}',
  },
  {
    key: "inputUnchanged",
    example: "select(foo)",
    jx: "map(select(. >= 2))",
    data: '{"a": 1, "b": 2, "c": 4, "d": 7}',
  },
  {
    key: "invokeFilter",
    example: "map(foo)",
    jx: "map(.+1)",
    data: '{"a": 1, "b": 2, "c": 3}',
  },
  {
    key: "conditionals",
    example: "if-then-else-end",
    jx: 'if .foo == 0 then "zero" elif .foo == 1 then "one" else "many" end',
    data: '{"foo": 2}',
  },
  {
    key: "stringInterpolation",
    example: "(foo)",
    jx: '"The input was \\(.input), which is one less than \\(.input+1)"',
    data: '{"input": 42}',
  },
] as const;
export default snippets;
