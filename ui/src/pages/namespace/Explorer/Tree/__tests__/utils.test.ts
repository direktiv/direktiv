import { describe, expect, test } from "vitest";
import {
  forceFileExtension,
  forceYamlFileExtension,
  stripFileExtension,
} from "../utils";

describe("forceYamlFileExtension", () => {
  test("it adds .yaml to a string that does not end on .yaml or yml", () => {
    expect(forceYamlFileExtension("somefile")).toBe("somefile.yaml");
  });

  test("it trims the input before adding a yaml", () => {
    expect(forceYamlFileExtension("somefile ")).toBe("somefile.yaml");
    expect(forceYamlFileExtension(" somefile")).toBe("somefile.yaml");
    expect(forceYamlFileExtension(" somefile ")).toBe("somefile.yaml");
  });

  test("it trims the input even when no extension is required", () => {
    expect(forceYamlFileExtension("somefile.yaml ")).toBe("somefile.yaml");
    expect(forceYamlFileExtension(" somefile.yaml")).toBe("somefile.yaml");
    expect(forceYamlFileExtension(" somefile.yaml ")).toBe("somefile.yaml");
  });

  test("it does nothing when the string ends with .yaml", () => {
    expect(forceYamlFileExtension("some-file.yaml")).toBe("some-file.yaml");
  });

  test("it does nothing when the string ends with .yml", () => {
    expect(forceYamlFileExtension("some-file.yml")).toBe("some-file.yml");
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
  test("it does nothing if the provided extension isn't contained in the file name", () => {
    expect(subject("name.ts", "notfound")).toBe("name.ts");
  });
});

describe("forceFileExtension", () => {
  const subject = forceFileExtension;
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
