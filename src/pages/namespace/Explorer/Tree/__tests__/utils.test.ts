import { describe, expect, test } from "vitest";

import { addYamlFileExtension } from "../utils";

describe("addYamlFileExtension", () => {
  test("it adds .yaml to a string that does not end on .yaml or yml", () => {
    expect(addYamlFileExtension("somefile")).toBe("somefile.yaml");
  });

  test("it trims the input before adding a yaml", () => {
    expect(addYamlFileExtension("somefile ")).toBe("somefile.yaml");
    expect(addYamlFileExtension(" somefile")).toBe("somefile.yaml");
    expect(addYamlFileExtension(" somefile ")).toBe("somefile.yaml");
  });

  test("it trims the input even when no extension is required", () => {
    expect(addYamlFileExtension("somefile.yaml ")).toBe("somefile.yaml");
    expect(addYamlFileExtension(" somefile.yaml")).toBe("somefile.yaml");
    expect(addYamlFileExtension(" somefile.yaml ")).toBe("somefile.yaml");
  });

  test("it will do nothing when the string ends with .yaml", () => {
    expect(addYamlFileExtension("some-file.yaml")).toBe("some-file.yaml");
  });

  test("it will do nothing when the string ends with .yml", () => {
    expect(addYamlFileExtension("some-file.yml")).toBe("some-file.yml");
  });
});
