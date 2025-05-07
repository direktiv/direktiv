import { describe, expect, test } from "vitest";
import { getValueFromJsonPath, parseVariable, variablePattern } from "../utils";

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
      src: "query.company-list.data.0.name",
      namespace: "query",
      id: "company-list",
      pointer: "data.0.name",
    });
  });

  test("it should handle unknown namespaces", () => {
    const result = parseVariable("unknown.company-list.data");
    expect(result).toEqual({
      src: "unknown.company-list.data",
      namespace: undefined,
      id: "company-list",
      pointer: "data",
    });
  });

  test("it should handle variables with just namespace and id", () => {
    const result = parseVariable("query.company-list");
    expect(result).toEqual({
      src: "query.company-list",
      namespace: "query",
      id: "company-list",
      pointer: undefined,
    });
  });

  test("it should handle variables with just namespace", () => {
    const result = parseVariable("query");
    expect(result).toEqual({
      src: "query",
      namespace: "query",
      id: undefined,
      pointer: undefined,
    });
  });

  test("it should handle empty variables", () => {
    const result = parseVariable("");
    expect(result).toEqual({
      src: "",
      namespace: undefined,
      id: undefined,
      pointer: undefined,
    });
  });
});

describe("getValueFromJsonPath", () => {
  describe("objects", () => {
    test("it should get values from a flat object", () => {
      expect(getValueFromJsonPath({ key: "value" }, "key")).toStrictEqual([
        "value",
        undefined,
      ]);
    });

    test("it should get values from nested objects", () => {
      const nestedObject = {
        user: {
          name: "John",
          address: { street: "123 Main St" },
        },
      };

      expect(getValueFromJsonPath(nestedObject, "user.name")).toStrictEqual([
        "John",
        undefined,
      ]);

      expect(
        getValueFromJsonPath(nestedObject, "user.address.street")
      ).toStrictEqual(["123 Main St", undefined]);
    });

    test("it should handle object keys that are numbers", () => {
      const obj = {
        "1": "one",
        nested: {
          2: "two",
        },
      };
      expect(getValueFromJsonPath(obj, "1")).toStrictEqual(["one", undefined]);
      expect(getValueFromJsonPath(obj, "nested.2")).toStrictEqual([
        "two",
        undefined,
      ]);
    });

    test("it will ignore keys that have dots in them", () => {
      const obj = {
        "some.key": "value",
        some: {
          key: "another value",
        },
      };

      expect(getValueFromJsonPath(obj, "some.key")).toStrictEqual([
        "another value",
        undefined,
      ]);
    });
  });

  describe("arrays", () => {
    test("it should get values from a flat array", () => {
      expect(getValueFromJsonPath(["value"], "0")).toStrictEqual([
        "value",
        undefined,
      ]);
    });

    test("it should get values from a nested array", () => {
      const oneLevelNestedArray = [["a", "b"]];
      const multiLevelNestedArray = [
        [
          ["a", "b"],
          ["c", "d", "e"],
        ],
      ];

      expect(getValueFromJsonPath(oneLevelNestedArray, "0.0")).toStrictEqual([
        "a",
        undefined,
      ]);

      expect(getValueFromJsonPath(oneLevelNestedArray, "0.1")).toStrictEqual([
        "b",
        undefined,
      ]);

      expect(
        getValueFromJsonPath(multiLevelNestedArray, "0.0.1")
      ).toStrictEqual(["b", undefined]);

      expect(
        getValueFromJsonPath(multiLevelNestedArray, "0.1.2")
      ).toStrictEqual(["e", undefined]);
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
      expect(getValueFromJsonPath(obj, "data.items.0.name")).toStrictEqual([
        "first",
        undefined,
      ]);
      expect(getValueFromJsonPath(obj, "data.items.1.id")).toStrictEqual([
        2,
        undefined,
      ]);
    });
  });

  test("it should accept an empty string to point to the root object", () => {
    expect(getValueFromJsonPath({ some: "object" }, "")).toStrictEqual([
      { some: "object" },
      undefined,
    ]);

    expect(getValueFromJsonPath(["some", "array"], "")).toStrictEqual([
      ["some", "array"],
      undefined,
    ]);
  });

  test("it should preserve the type of the value", () => {
    const obj = {
      string: "value",
      true: true,
      false: false,
      number: 42,
      zero: 0,
      null: null,
      array: ["value"],
      object: { key: "value" },
    };

    expect(getValueFromJsonPath(obj, "string")).toStrictEqual([
      "value",
      undefined,
    ]);

    expect(getValueFromJsonPath(obj, "true")).toStrictEqual([true, undefined]);

    expect(getValueFromJsonPath(obj, "false")).toStrictEqual([
      false,
      undefined,
    ]);

    expect(getValueFromJsonPath(obj, "number")).toStrictEqual([42, undefined]);

    expect(getValueFromJsonPath(obj, "zero")).toStrictEqual([0, undefined]);

    expect(getValueFromJsonPath(obj, "null")).toStrictEqual([null, undefined]);

    expect(getValueFromJsonPath(obj, "array")).toStrictEqual([
      ["value"],
      undefined,
    ]);

    expect(getValueFromJsonPath(obj, "object")).toStrictEqual([
      { key: "value" },
      undefined,
    ]);
  });

  describe("invalid data", () => {
    test("it should return an invalidPath error when the path does not exist", () => {
      const obj = { some: "object" };
      expect(getValueFromJsonPath(obj, "invalid.path")).toStrictEqual([
        undefined,
        "invalidPath",
      ]);
    });

    test("it should return an invalidPath error when the path points to an undefined value", () => {
      expect(
        getValueFromJsonPath({ undefinedValue: undefined }, "undefinedValue")
      ).toStrictEqual([undefined, "invalidPath"]);
    });

    test("it should return an invalidPath error when trying to point to a key with a dot in it", () => {
      expect(
        getValueFromJsonPath({ "some.path": "value" }, "some.path")
      ).toStrictEqual([undefined, "invalidPath"]);
    });

    test("it should return an invalidJson error when the input is  not a JSON", () => {
      expect(getValueFromJsonPath(false, "some.key")).toStrictEqual([
        undefined,
        "invalidJson",
      ]);

      expect(getValueFromJsonPath(true, "some.key")).toStrictEqual([
        undefined,
        "invalidJson",
      ]);

      expect(getValueFromJsonPath(undefined, "some.key")).toStrictEqual([
        undefined,
        "invalidJson",
      ]);

      expect(getValueFromJsonPath(null, "some.key")).toStrictEqual([
        undefined,
        "invalidJson",
      ]);

      expect(getValueFromJsonPath("string", "some.key")).toStrictEqual([
        undefined,
        "invalidJson",
      ]);

      expect(getValueFromJsonPath("", "some.key")).toStrictEqual([
        undefined,
        "invalidJson",
      ]);

      expect(getValueFromJsonPath(1, "some.key")).toStrictEqual([
        undefined,
        "invalidJson",
      ]);
    });
  });
});
