import {
  PermisionSchemaType,
  PermissionMethodAvailableUi,
  PermissionTopic,
} from "~/api/enterprise/tokens/schema";
import { describe, expect, test } from "vitest";

import { updatePermissions } from "../utils";

describe("updatePermissions", () => {
  const createPermission = (
    topic: PermissionTopic,
    method: PermissionMethodAvailableUi
  ): PermisionSchemaType => ({
    topic,
    method,
  });

  test("should add a new permission when none exists for the topic", () => {
    const initialPermissions: PermisionSchemaType[] = [];
    const result = updatePermissions({
      permissions: initialPermissions,
      topic: "namespaces",
      value: "read",
    });

    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({
      topic: "namespaces",
      method: "read",
    });
  });

  test("should update an existing permission", () => {
    const initialPermissions: PermisionSchemaType[] = [
      createPermission("namespaces", "read"),
    ];

    const result = updatePermissions({
      permissions: initialPermissions,
      topic: "namespaces",
      value: "manage",
    });

    expect(result).toHaveLength(1);
    expect(result[0]).toEqual({
      topic: "namespaces",
      method: "manage",
    });
  });

  test("should remove a permission when value is undefined", () => {
    const initialPermissions: PermisionSchemaType[] = [
      createPermission("namespaces", "read"),
      createPermission("instances", "manage"),
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

  test("should not modify other permissions when updating one", () => {
    const initialPermissions: PermisionSchemaType[] = [
      createPermission("namespaces", "read"),
      createPermission("instances", "manage"),
      createPermission("secrets", "read"),
    ];

    const result = updatePermissions({
      permissions: initialPermissions,
      topic: "instances",
      value: "read",
    });

    expect(result).toHaveLength(3);
    expect(result).toContainEqual(createPermission("namespaces", "read"));
    expect(result).toContainEqual(createPermission("instances", "read"));
    expect(result).toContainEqual(createPermission("secrets", "read"));
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
      createPermission("namespaces", "read"),
      createPermission("instances", "manage"),
      createPermission("secrets", "read"),
    ];

    const result = updatePermissions({
      permissions: initialPermissions,
      topic: "instances",
      value: "read",
    });

    expect(result).toEqual([
      createPermission("namespaces", "read"),
      createPermission("instances", "read"),
      createPermission("secrets", "read"),
    ]);
  });

  test("should create a new array and not modify the original", () => {
    const initialPermissions: PermisionSchemaType[] = [
      createPermission("namespaces", "read"),
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
