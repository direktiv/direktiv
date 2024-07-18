import {
  addFileExtension,
  addYamlFileExtension,
  stripFileExtension,
} from "../utils";
import { describe, expect, test } from "vitest";

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

  test("it does nothing when the string ends with .yaml", () => {
    expect(addYamlFileExtension("some-file.yaml")).toBe("some-file.yaml");
  });

  test("it does nothing when the string ends with .yml", () => {
    expect(addYamlFileExtension("some-file.yml")).toBe("some-file.yml");
  });
});

describe("stripFileExtension", () => {
  const subject = stripFileExtension;
  test("it strips the provided extension from the name", () => {
    expect(subject("name.foo.ts", "foo.ts")).toBe("name");
  });
  test("it strips the last segment from the file name if it matches a part of the provided extension", () => {
    expect(subject("name.ts", "foo.ts")).toBe("name");
    expect(subject("name.foo", "foo.ts")).toBe("name");
  });
});

describe("addFileExtension", () => {
  const subject = addFileExtension;
  const extension = ".workflow.ts";

  test("it adds the extension .workflow.ts", () => {
    expect(subject("some", extension)).toBe("some.workflow.ts");
  });

  test("it does not duplicate existing partial extensions", () => {
    expect(subject("some.ts", extension)).toBe("some.workflow.ts");
    expect(subject("some.workflow", extension)).toBe("some.workflow.ts");
    expect(subject("some.workflow.ts", extension)).toBe("some.workflow.ts");
  });

  test("it preserves segments separated with a dot", () => {
    expect(subject("some.name", extension)).toBe("some.name.workflow.ts");
    expect(subject("some.name.ts", extension)).toBe("some.name.workflow.ts");
    expect(subject("some.name.ts", extension)).toBe("some.name.workflow.ts");
    expect(subject("some.name.workflow", extension)).toBe(
      "some.name.workflow.ts"
    );
  });

  test("it trims the input before adding the extension .workflow.ts ", () => {
    expect(subject("some ", extension)).toBe("some.workflow.ts");
    expect(subject(" some", extension)).toBe("some.workflow.ts");
    expect(subject(" some ", extension)).toBe("some.workflow.ts");
  });

  test("it trims the input even when extension already exists", () => {
    expect(subject("some.workflow.ts ", extension)).toBe("some.workflow.ts");
    expect(subject(" some.workflow.ts", extension)).toBe("some.workflow.ts");
    expect(subject(" some.workflow.ts ", extension)).toBe("some.workflow.ts");
  });
});
