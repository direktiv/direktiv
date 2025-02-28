import { describe, expect, test } from "vitest";

import { PermisionSchemaType } from "~/api/enterprise/schema";
import { updatePermissions } from "../utils";

describe("updatePermissions", () => {
  test("should add a new permission when none exists for the topic", () => {
    const initialPermissions: PermisionSchemaType[] = [
      {
        topic: "events",
        method: "manage",
      },
    ];
    const result = updatePermissions({
      permissions: initialPermissions,
      topic: "namespaces",
      value: "read",
    });

    expect(result).toHaveLength(2);
    expect(result[1]).toEqual({
      topic: "namespaces",
      method: "read",
    });
  });

  test("it should update an existing permission", () => {
    const initialPermissions: PermisionSchemaType[] = [
      { topic: "namespaces", method: "read" },
      { topic: "instances", method: "manage" },
      { topic: "secrets", method: "read" },
    ];

    const result = updatePermissions({
      permissions: initialPermissions,
      topic: "instances",
      value: "read",
    });

    expect(result).toHaveLength(3);
    expect(result).toContainEqual({ topic: "namespaces", method: "read" });
    expect(result).toContainEqual({ topic: "instances", method: "read" });
    expect(result).toContainEqual({ topic: "secrets", method: "read" });
  });

  test("should remove a permission when value is undefined", () => {
    const initialPermissions: PermisionSchemaType[] = [
      { topic: "namespaces", method: "read" },
      { topic: "instances", method: "manage" },
    ];

    const result = updatePermissions({
      permissions: initialPermissions,
      topic: "namespaces",
      value: undefined,
    });

    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({
      topic: "instances",
      method: "manage",
    });
  });

  test("should handle empty permissions array", () => {
    const result = updatePermissions({
      permissions: [],
      topic: "variables",
      value: "manage",
    });

    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({
      topic: "variables",
      method: "manage",
    });
  });

  test("should maintain permission order when updating", () => {
    const initialPermissions: PermisionSchemaType[] = [
      { topic: "namespaces", method: "read" },
      { topic: "instances", method: "manage" },
      { topic: "secrets", method: "read" },
    ];

    const result = updatePermissions({
      permissions: initialPermissions,
      topic: "instances",
      value: "read",
    });

    expect(result).toEqual([
      { topic: "namespaces", method: "read" },
      { topic: "instances", method: "read" },
      { topic: "secrets", method: "read" },
    ]);
  });

  test("should create a new array and not modify the original", () => {
    const initialPermissions: PermisionSchemaType[] = [
      { topic: "namespaces", method: "read" },
    ];

    const result = updatePermissions({
      permissions: initialPermissions,
      topic: "namespaces",
      value: "manage",
    });

    expect(result).not.toBe(initialPermissions);
    expect(initialPermissions[0]?.method).toBe("read");
  });
});
