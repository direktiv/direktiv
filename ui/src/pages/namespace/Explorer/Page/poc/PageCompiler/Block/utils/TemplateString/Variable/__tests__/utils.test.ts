import { describe, expect, test } from "vitest";
import { parseVariable, variablePattern } from "../utils";

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

  test("it is valid to use {{ or }}outside of a variable", () => {
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
