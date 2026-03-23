import { describe, expect, test } from "vitest";

import {
  isActionEntityTypeName,
  isEntityOrCommonName,
  isEntityTypeName,
  isExtensionTypeName,
  isIdentifierPath,
  isNamespaceName,
  isSchemaTypeReferenceName,
} from "../identifiers";

describe("isIdentifierPath", () => {
  test("accepts valid Cedar identifier paths", () => {
    expect(isIdentifierPath("User")).toBe(true);
    expect(isIdentifierPath("MyNamespace::User")).toBe(true);
    expect(isIdentifierPath("_internal::User_1")).toBe(true);
  });

  test("rejects empty values by default", () => {
    expect(isIdentifierPath("")).toBe(false);
  });

  test("accepts the empty value when allowEmpty is enabled", () => {
    expect(isIdentifierPath("", { allowEmpty: true })).toBe(true);
  });

  test("rejects invalid identifier segments", () => {
    expect(isIdentifierPath("1User")).toBe(false);
    expect(isIdentifierPath("User::invalid-name")).toBe(false);
    expect(isIdentifierPath("User::")).toBe(false);
  });

  test("rejects reserved Cedar namespace segments by default", () => {
    expect(isIdentifierPath("__cedar::ipaddr")).toBe(false);
    expect(isIdentifierPath("App::__cedar::Type")).toBe(false);
  });

  test("accepts reserved Cedar namespace segments when enabled", () => {
    expect(
      isIdentifierPath("__cedar::ipaddr", {
        allowReservedCedarNamespace: true,
      })
    ).toBe(true);
  });

  test("rejects reserved Cedar type names in the final segment by default", () => {
    expect(isIdentifierPath("String")).toBe(false);
    expect(isIdentifierPath("MyNamespace::Record")).toBe(false);
  });

  test("accepts reserved final segments when enabled", () => {
    expect(
      isIdentifierPath("__cedar::ipaddr", {
        allowReservedCedarNamespace: true,
        allowReservedFinalSegment: true,
      })
    ).toBe(true);

    expect(
      isIdentifierPath("Shared::Action", {
        allowReservedFinalSegment: true,
      })
    ).toBe(true);
  });
});

describe("identifier helpers", () => {
  test("isNamespaceName accepts the empty namespace and qualified names", () => {
    expect(isNamespaceName("")).toBe(true);
    expect(isNamespaceName("Demo::Namespace")).toBe(true);
    expect(isNamespaceName("__cedar")).toBe(false);
  });

  test("isEntityTypeName rejects reserved and malformed names", () => {
    expect(isEntityTypeName("User")).toBe(true);
    expect(isEntityTypeName("String")).toBe(false);
    expect(isEntityTypeName("invalid-name")).toBe(false);
  });

  test("isEntityOrCommonName allows reserved final segments and Cedar namespace types", () => {
    expect(isEntityOrCommonName("User")).toBe(true);
    expect(isEntityOrCommonName("String")).toBe(true);
    expect(isEntityOrCommonName("__cedar::ipaddr")).toBe(true);
    expect(isEntityOrCommonName("invalid-name")).toBe(false);
  });

  test("isExtensionTypeName follows Cedar extension naming rules", () => {
    expect(isExtensionTypeName("ipaddr")).toBe(true);
    expect(isExtensionTypeName("__cedar::decimal")).toBe(true);
    expect(isExtensionTypeName("__cedar::invalid-name")).toBe(false);
  });

  test("isActionEntityTypeName only accepts Action entity types", () => {
    expect(isActionEntityTypeName("Action")).toBe(true);
    expect(isActionEntityTypeName("Shared::Action")).toBe(true);
    expect(isActionEntityTypeName("Shared::NotAction")).toBe(false);
    expect(isActionEntityTypeName("invalid-name")).toBe(false);
  });

  test("isSchemaTypeReferenceName excludes Cedar keywords", () => {
    expect(isSchemaTypeReferenceName("CustomType")).toBe(true);
    expect(isSchemaTypeReferenceName("Long")).toBe(false);
    expect(isSchemaTypeReferenceName("EntityOrCommon")).toBe(false);
    expect(isSchemaTypeReferenceName("__cedar::ipaddr")).toBe(true);
  });
});
