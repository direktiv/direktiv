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

  test("rejects invalid identifier paths by default", () => {
    expect(isIdentifierPath("")).toBe(false);
    expect(isIdentifierPath("1User")).toBe(false);
    expect(isIdentifierPath("invalid-name")).toBe(false);
    expect(isIdentifierPath("User::invalid-name")).toBe(false);
    expect(isIdentifierPath("User::")).toBe(false);
    expect(isIdentifierPath("__cedar::ipaddr")).toBe(false);
    expect(isIdentifierPath("App::__cedar::Type")).toBe(false);
    expect(isIdentifierPath("String")).toBe(false);
    expect(isIdentifierPath("MyNamespace::Record")).toBe(false);
  });

  test("accepts allowed special cases when options are enabled", () => {
    expect(isIdentifierPath("", { allowEmpty: true })).toBe(true);

    expect(
      isIdentifierPath("__cedar::ipaddr", { allowReservedCedarNamespace: true })
    ).toBe(true);

    expect(
      isIdentifierPath("__cedar::ipaddr", {
        allowReservedCedarNamespace: true,
        allowReservedFinalSegment: true,
      })
    ).toBe(true);

    expect(
      isIdentifierPath("Shared::Action", { allowReservedFinalSegment: true })
    ).toBe(true);
  });
});

describe("isNamespaceName", () => {
  test("accepts valid namespace names", () => {
    expect(isNamespaceName("")).toBe(true);
    expect(isNamespaceName("Demo::Namespace")).toBe(true);
  });

  test("rejects invalid namespace names", () => {
    expect(isNamespaceName("__cedar")).toBe(false);
  });
});

describe("isEntityTypeName", () => {
  test("accepts valid entity type names", () => {
    expect(isEntityTypeName("User")).toBe(true);
  });

  test("rejects reserved and malformed entity type names", () => {
    expect(isEntityTypeName("String")).toBe(false);
    expect(isEntityTypeName("invalid-name")).toBe(false);
  });
});

describe("isEntityOrCommonName", () => {
  test("accepts valid entity or common names", () => {
    expect(isEntityOrCommonName("User")).toBe(true);
    expect(isEntityOrCommonName("String")).toBe(true);
    expect(isEntityOrCommonName("__cedar::ipaddr")).toBe(true);
  });

  test("rejects invalid entity or common names", () => {
    expect(isEntityOrCommonName("invalid-name")).toBe(false);
  });
});

describe("isExtensionTypeName", () => {
  test("accepts valid extension type names", () => {
    expect(isExtensionTypeName("ipaddr")).toBe(true);
    expect(isExtensionTypeName("__cedar::decimal")).toBe(true);
  });

  test("rejects invalid extension type names", () => {
    expect(isExtensionTypeName("__cedar::invalid-name")).toBe(false);
  });
});

describe("isActionEntityTypeName", () => {
  test("accepts valid action entity type names", () => {
    expect(isActionEntityTypeName("Action")).toBe(true);
    expect(isActionEntityTypeName("Shared::Action")).toBe(true);
  });

  test("rejects invalid action entity type names", () => {
    expect(isActionEntityTypeName("Shared::NotAction")).toBe(false);
    expect(isActionEntityTypeName("invalid-name")).toBe(false);
  });
});

describe("isSchemaTypeReferenceName", () => {
  test("accepts valid schema type references", () => {
    expect(isSchemaTypeReferenceName("CustomType")).toBe(true);
    expect(isSchemaTypeReferenceName("__cedar::ipaddr")).toBe(true);
  });

  test("rejects Cedar keywords as schema type references", () => {
    expect(isSchemaTypeReferenceName("Long")).toBe(false);
    expect(isSchemaTypeReferenceName("EntityOrCommon")).toBe(false);
  });
});
