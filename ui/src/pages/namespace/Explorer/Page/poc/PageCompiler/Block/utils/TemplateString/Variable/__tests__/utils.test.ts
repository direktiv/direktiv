import { describe, expect, test } from "vitest";
import { getObjectValueByPath, parseVariable, variablePattern } from "../utils";

describe("Template string variable regex", () => {
  test("it should match the basic variable syntax", () => {
    const template = "{{name}}";
    const matches = Array.from(template.matchAll(variablePattern));

    expect(matches.length).toBe(1);
    expect(matches[0]?.[0]).toBe("{{name}}");
    expect(matches[0]?.[1]).toBe("name");
  });

  test("it should match variables with whitespace", () => {
    const template = "{{ query }}";
    const matches = Array.from(template.matchAll(variablePattern));

    expect(matches.length).toBe(1);
    expect(matches[0]?.[0]).toBe("{{ query }}");
    expect(matches[0]?.[1]).toBe("query");
  });

  test("it should match variables with inconsistent whitespace", () => {
    const template = "{{  query}}";
    const matches = Array.from(template.matchAll(variablePattern));

    expect(matches.length).toBe(1);
    expect(matches[0]?.[0]).toBe("{{  query}}");
    expect(matches[0]?.[1]).toBe("query");

    const template2 = "{{query   }}";
    const matches2 = Array.from(template2.matchAll(variablePattern));

    expect(matches2.length).toBe(1);
    expect(matches2[0]?.[0]).toBe("{{query   }}");
    expect(matches2[0]?.[1]).toBe("query");
  });

  test("it should match variable names with special characters", () => {
    const template = "{{ query.pokemon-list.data.0.name}}";
    const matches = Array.from(template.matchAll(variablePattern));

    expect(matches.length).toBe(1);
    expect(matches[0]?.[0]).toBe("{{ query.pokemon-list.data.0.name}}");
    expect(matches[0]?.[1]).toBe("query.pokemon-list.data.0.name");
  });

  test("it should find multiple variables in a template", () => {
    const template =
      "Hello {{query.user-data.data.salutation}} {{query.user-data.data.name}}";
    const matches = Array.from(template.matchAll(variablePattern));

    expect(matches.length).toBe(2);
    expect(matches[0]?.[1]).toBe("query.user-data.data.salutation");
    expect(matches[1]?.[1]).toBe("query.user-data.data.name");
  });

  test("it should handle variables diretly next to each other", () => {
    const template = "{{first}}{{second}}";
    const matches = Array.from(template.matchAll(variablePattern));

    expect(matches.length).toBe(2);
    expect(matches[0]?.[1]).toBe("first");
    expect(matches[1]?.[1]).toBe("second");
  });

  test("it should not match incomplete variable syntax", () => {
    const template = "{{ name";
    const matches = Array.from(template.matchAll(variablePattern));

    expect(matches.length).toBe(0);

    const template2 = "name }}";
    const matches2 = Array.from(template2.matchAll(variablePattern));

    expect(matches2.length).toBe(0);
  });

  test("it should not match variables that contains curly braces", () => {
    const template = "{{ novar} }} {{ {novar }}";
    const matches = Array.from(template.matchAll(variablePattern));
    expect(matches.length).toBe(0);
  });

  test("it is valid to use {{ or }} outside of a variable", () => {
    const template =
      "Hello {{ name }}, these are not variables: {{}}, }}  and {{";
    const matches = Array.from(template.matchAll(variablePattern));

    expect(matches.length).toBe(1);
    expect(matches[0]?.[0]).toBe("{{ name }}");
    expect(matches[0]?.[1]).toBe("name");
  });
});

describe("parseVariable", () => {
  test("it should parse a complete variable with namespace, id and pointer", () => {
    const result = parseVariable("query.company-list.data.0.name");
    expect(result).toEqual({
      namespace: "query",
      id: "company-list",
      pointer: "data.0.name",
    });
  });

  test("it should handle unknown namespaces", () => {
    const result = parseVariable("unknown.company-list.data");
    expect(result).toEqual({
      namespace: undefined,
      id: "company-list",
      pointer: "data",
    });
  });

  test("it should handle variables with just namespace and id", () => {
    const result = parseVariable("query.company-list");
    expect(result).toEqual({
      namespace: "query",
      id: "company-list",
      pointer: undefined,
    });
  });

  test("it should handle variables with just namespace", () => {
    const result = parseVariable("query");
    expect(result).toEqual({
      namespace: "query",
      id: undefined,
      pointer: undefined,
    });
  });

  test("it should handle emopty variables", () => {
    const result = parseVariable("");
    expect(result).toEqual({
      namespace: undefined,
      id: undefined,
      pointer: undefined,
    });
  });
});

describe("getObjectValueByPath", () => {
  describe("objects", () => {
    test("it should get values from a flat object", () => {
      expect(getObjectValueByPath({ key: "value" }, "key")).toBe("value");
    });

    test("it should get values from nested objects", () => {
      const nestedObject = {
        user: {
          name: "John",
          address: { street: "123 Main St" },
        },
      };

      expect(getObjectValueByPath(nestedObject, "user.name")).toBe("John");
      expect(getObjectValueByPath(nestedObject, "user.address.street")).toBe(
        "123 Main St"
      );
    });

    test("it address keys that are numbers", () => {
      const obj = {
        "1": "one",
        nested: {
          2: "two",
        },
      };
      expect(getObjectValueByPath(obj, "1")).toBe("one");
      expect(getObjectValueByPath(obj, "nested.2")).toBe("two");
    });
  });

  describe("arrays", () => {
    test("it should get values from a flat array", () => {
      expect(getObjectValueByPath(["value"], "0")).toBe("value");
    });

    test("it should get values from a nested array", () => {
      const oneLevelNestedArray = [["a", "b"]];
      const multiLevelNestedArray = [
        [
          ["a", "b"],
          ["c", "d", "e"],
        ],
      ];

      expect(getObjectValueByPath(oneLevelNestedArray, "0.0")).toBe("a");
      expect(getObjectValueByPath(oneLevelNestedArray, "0.1")).toBe("b");
      expect(getObjectValueByPath(multiLevelNestedArray, "0.0.1")).toBe("b");
      expect(getObjectValueByPath(multiLevelNestedArray, "0.1.2")).toBe("e");
    });

    test("it should handle array indices in the path", () => {
      const obj = {
        data: {
          items: [
            { id: 1, name: "first" },
            { id: 2, name: "second" },
          ],
        },
      };
      expect(getObjectValueByPath(obj, "data.items.0.name")).toBe("first");
      expect(getObjectValueByPath(obj, "data.items.1.id")).toBe("2");
    });
  });

  describe("type casting", () => {
    test("it should stringify numbers and booleans", () => {
      const obj = {
        true: true,
        false: false,
        number: 42,
        zero: 0,
      };
      expect(getObjectValueByPath(obj, "true")).toBe("true");
      expect(getObjectValueByPath(obj, "false")).toBe("false");
      expect(getObjectValueByPath(obj, "number")).toBe("42");
      expect(getObjectValueByPath(obj, "zero")).toBe("0");
    });

    test("it should return undefined if values are null or undefined", () => {
      const obj = {
        nullValue: null,
        undefinedValue: undefined,
        nested: { nullValue: null },
      };
      expect(getObjectValueByPath(obj, "nullValue")).toBe(undefined);
      expect(getObjectValueByPath(obj, "undefinedValue")).toBe(undefined);
      expect(getObjectValueByPath(obj, "nested.nullValue")).toBe(undefined);
    });

    test("it should return <Array> if value is in array", () => {
      const obj = {
        array: { empty: [], nonEmpty: ["value"] },
      };

      expect(getObjectValueByPath(obj, "array.empty")).toBe("<Array>");
      expect(getObjectValueByPath(obj, "array.nonEmpty")).toBe("<Array>");
    });

    test("it should return <Object> if value is in object", () => {
      const obj = {
        object: { empty: {}, nonEmpty: { key: "value" } },
      };

      expect(getObjectValueByPath(obj, "object.empty")).toBe("<Object>");
      expect(getObjectValueByPath(obj, "object.nonEmpty")).toBe("<Object>");
    });
  });

  describe("invalid data", () => {
    test("it should return undefined for invalid paths", () => {
      const obj = { data: { name: "test" } };
      expect(getObjectValueByPath(obj, "invalid.path")).toBe(undefined);
      expect(getObjectValueByPath(obj, "")).toBe(undefined);
    });

    test("it should return undefined for invalid inputs", () => {
      expect(getObjectValueByPath(false, "some.key")).toBe(undefined);
      expect(getObjectValueByPath(true, "some.key")).toBe(undefined);
      expect(getObjectValueByPath(undefined, "some.key")).toBe(undefined);
      expect(getObjectValueByPath(null, "some.key")).toBe(undefined);
      expect(getObjectValueByPath("string", "some.key")).toBe(undefined);
      expect(getObjectValueByPath("", "some.key")).toBe(undefined);
      expect(getObjectValueByPath(1, "some.key")).toBe(undefined);
    });
  });
});
