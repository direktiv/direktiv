import * as React from 'react'
const fetch = require('isomorphic-fetch')
import { HandleError, ExtractQueryString, apiKeyHeaders } from '../util'


const cheatSheetMap = [
    {
        example: ".",
        tip: "unchanged input",
        filter: ".",
        json: '{ "foo": { "bar": { "baz": 123 } } }',
    },
    {
        example: ".foo, .foo.bar, .foo?",
        tip: "value at key",
        filter: ".foo",
        json: '{"foo": 42, "bar": "less interesting data"}',
    },
    {
        example: ".[], .[]?, .[2], .[10:15]",
        tip: "array operation",
        filter: ".foo[1]",
        json:
            '{"foo": [{"name":"JSON", "good":true}, {"name":"XML", "good":false}]}',
    },
    {
        example: "[], {}",
        tip: "array/object construction",
        filter: "{user, title: .titles[]}",
        json: '{"user":"stedolan","titles":["JQ Primer", "More JQ"]}',
    },
    {
        example: "length",
        tip: "length of a value",
        filter: ".foo[] | length",
        json: '{"foo": [[1,2], "string", {"a":2}, null]}',
    },
    {
        example: "keys",
        tip: "keys in an array",
        filter: "keys",
        json: '{"abc": 1, "abcd": 2, "Foo": 3}',
    },
    {
        example: ",",
        tip: "feed input into multiple filters",
        filter: ".foo, .bar",
        json: '{ "foo": 42, "bar": "something else", "baz": true}',
    },
    {
        example: "|",
        tip: "pipe output of one filter to the next filter",
        filter: ".foo[] | .name",
        json:
            '{"foo": [{"name":"JSON", "good":true}, {"name":"XML", "good":false}]}',
    },
    {
        example: "select(foo)",
        tip: "input unchanged if foo returns true",
        filter: "map(select(. >= 2))",
        json: '{"a": 1, "b": 2, "c": 4, "d": 7}',
    },
    {
        example: "map(foo)",
        tip: "invoke filter foo for each input",
        filter: "map(.+1)",
        json: '{"a": 1, "b": 2, "c": 3}',
    },
    {
        example: "if-then-else-end",
        tip: "conditionals",
        filter:
            'if .foo == 0 then "zero" elif .foo == 1 then "one" else "many" end',
        json: '{"foo": 2}',
    },
    {
        example: "(foo)",
        tip: "string interpolation",
        filter: '"The input was \\(.input), which is one less than \\(.input+1)"',
        json: '{"input": 42}',
    },
];

/*
    useJQPlayground is a react hook which returns createNamespace, deleteNamespace and data
    takes:
      - url to direktiv api http://x/api/
      - apikey to provide authentication of an apikey
*/
export const useDirektivJQPlayground = (url, apikey) => {

    const [data, setData] = React.useState(null)

    async function executeJQ(query, data, ...queryParameters) {
        // fetch namespace list by default
        let resp = await fetch(`${url}jq${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey),
            method: "POST",
            body: JSON.stringify({
                query: query,
                data: data,
            })
        })
        if (!resp.ok) {
            throw new Error((await HandleError('execute jq', resp, 'jqPlayground')))
        }

        let json = await resp.json()
        setData(json.results)
        return json.results
    }


    return {
        data,
        executeJQ,
        cheatSheet: cheatSheetMap,
    }
}