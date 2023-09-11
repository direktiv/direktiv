import { addYamlFileExtension, removeYamlFileExtension } from "../utils";
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

  test("it will do nothing when the string ends with .yaml", () => {
    expect(addYamlFileExtension("some-file.yaml")).toBe("some-file.yaml");
  });

  test("it will do nothing when the string ends with .yml", () => {
    expect(addYamlFileExtension("some-file.yml")).toBe("some-file.yml");
  });
});

describe("removeYamlFileExtension", () => {
  test("it removes .yaml from a string that end on .yaml", () => {
    expect(removeYamlFileExtension("somefile.yaml")).toBe("somefile");
  });

  test("it removes .yml from a string that end on .yml", () => {
    expect(removeYamlFileExtension("somefile.yml")).toBe("somefile");
  });

  test("it will trim the string before removing the extension", () => {
    expect(removeYamlFileExtension("somefile.yml ")).toBe("somefile");
    expect(removeYamlFileExtension("somefile.yaml ")).toBe("somefile");
  });

  test("it will not change other file extensions", () => {
    expect(removeYamlFileExtension("somefile.jpg")).toBe("somefile.jpg");
    expect(removeYamlFileExtension("somefile.ymly")).toBe("somefile.ymly");
  });

  test("it will not remove anything from files without any file extension", () => {
    expect(removeYamlFileExtension("somefile")).toBe("somefile");
    expect(removeYamlFileExtension("somefileendingwithyaml")).toBe(
      "somefileendingwithyaml"
    );
    expect(removeYamlFileExtension("somefileendingwithyml")).toBe(
      "somefileendingwithyml"
    );
  });

  test("it will trim the filename, even if no extension was found", () => {
    expect(removeYamlFileExtension("somefile ")).toBe("somefile");
    expect(removeYamlFileExtension("somefile.jpg ")).toBe("somefile.jpg");
  });
});
