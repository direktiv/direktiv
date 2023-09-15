type SnippetKey =
  | "unchangedInput"
  | "valueAtKey"
  | "arrayOperation"
  | "arrayObjectConstruction"
  | "arrayObjectConstruction"
  | "lengthOfValue"
  | "keysInArray"
  | "feedInput"
  | "pipeOutput"
  | "inputUnchanged"
  | "invokeFilter"
  | "conditionals"
  | "stringInterpolation";

export interface Snippet {
  key: SnippetKey;
  example: string;
  query: string;
  input: string;
  output:
    | {
        [key: string]: any;
      }
    | string
    | number;
}

const snippets: Snippet[] = [
  {
    key: "unchangedInput",
    example: ".",
    query: ".",
    input: '{ "foo": { "bar": { "baz": 123 } } }',
    output: {
      foo: {
        bar: {
          baz: 123,
        },
      },
    },
  },
  {
    key: "valueAtKey",
    example: ".foo, .foo.bar, .foo?",
    query: ".foo",
    input: '{"foo": 42, "bar": "less interesting data"}',
    output: "42",
  },
  {
    key: "arrayOperation",
    example: ".[], .[]?, .[2], .[10:15]",
    query: ".foo[1]",
    input:
      '{"foo": [{"name":"JSON", "good":true}, {"name":"XML", "good":false}]}',
    output: {
      good: false,
      name: "XML",
    },
  },
  {
    key: "arrayObjectConstruction",
    example: "[], {}",
    query: "[{user, title: .titles[]}]",
    input: '{"user":"stedolan","titles":["JQ Primer", "More JQ"]}',
    output: [
      {
        title: "JQ Primer",
        user: "stedolan",
      },
      {
        title: "More JQ",
        user: "stedolan",
      },
    ],
  },
  {
    key: "lengthOfValue",
    example: "length",
    query: "[.foo[] | length]",
    input: '{"foo": [[1,2], "string", {"a":2}, null]}',
    output: [2, 6, 1, 0],
  },
  {
    key: "keysInArray",
    example: "keys",
    query: "keys",
    input: '{"abc": 1, "abcd": 2, "Foo": 3}',
    output: ["Foo", "abc", "abcd"],
  },
  {
    key: "feedInput",
    example: ",",
    query: "[.foo, .bar]",
    input: '{ "foo": 42, "bar": "something else", "baz": true}',
    output: [42, "something else"],
  },
  {
    key: "pipeOutput",
    example: "|",
    query: "[.foo[] | .name]",
    input:
      '{"foo": [{"name":"JSON", "good":true}, {"name":"XML", "good":false}]}',
    output: ["JSON", "XML"],
  },
  {
    key: "inputUnchanged",
    example: "select(foo)",
    query: "map(select(. >= 2))",
    input: '{"a": 1, "b": 2, "c": 4, "d": 7}',
    output: [2, 4, 7],
  },
  {
    key: "invokeFilter",
    example: "map(foo)",
    query: "map(.+1)",
    input: '{"a": 1, "b": 2, "c": 3}',
    output: [2, 3, 4],
  },
  {
    key: "conditionals",
    example: "if-then-else-end",
    query: 'if .foo == 0 then "zero" elif .foo == 1 then "one" else "many" end',
    input: '{"foo": 2}',
    output: '"many"',
  },
  {
    key: "stringInterpolation",
    example: "(foo)",
    query: '"The input was \\(.input), which is one less than \\(.input+1)"',
    input: '{"input": 42}',
    output: '"The input was 42, which is one less than 43"',
  },
] as const;
export default snippets;
