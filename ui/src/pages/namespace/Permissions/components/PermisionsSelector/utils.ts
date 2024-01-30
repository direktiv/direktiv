/**
 * takes a permission string and returns the scope and resource
 *
 * example input: "READ:config"
 *
 * example output: ["READ", "config"]
 */
const splitPermissionString = (permission: string): [string, string] => {
  const [scope, resource] = permission.split(":");
  return [scope ?? "", resource ?? ""];
};

/**
 * takes a scope and resource and returns a permission string
 *
 * example input: ["READ", "config"]
 *
 * example output: "READ:config"
 *
 */
export const joinPermissionString = (scope: string, resource: string): string =>
  `${scope}:${resource}`;

/**
 * takes a list of permission strings and returns all availables of scopes
 *
 * example input:
 * [
 *   "READ:config",
 *   "WRITE:config",
 *   "READ:lint",
 *   "WRITE:lint",
 *   "READ:logs",
 *   "WRITE:logs"
 * ]
 *
 * example output:
 * [
 *   "READ",
 *   "WRITE"
 * ]
 *
 */
export const permissionStringsToScopes = (permissions: string[]) =>
  permissions.reduce((allScopes, permissionString) => {
    const [scope] = splitPermissionString(permissionString);
    if (allScopes.includes(scope)) {
      return allScopes;
    }

    return [...allScopes, scope];
  }, [] as string[]);

/**
 * takes a list of permission strings and groups them by resource
 *
 * example input:
 * [
 *   "READ:config",
 *   "WRITE:config",
 *   "READ:lint",
 *   "WRITE:lint",
 *   "READ:logs",
 *   "WRITE:logs"
 * ]
 *
 * example output:
 * {
 *   config: ["READ", "WRITE"],
 *   lint: ["READ", "WRITE"],
 *   logs: ["READ", "WRITE"]
 * }
 *
 */
type GroupedPermission = Record<string, string[]>;
export const groupPermissionStringsByResouce = (permissions: string[]) =>
  permissions.reduce<GroupedPermission>(
    (groupedResources, permissionString) => {
      const [scope, resource] = splitPermissionString(permissionString);
      const existingEntries = groupedResources[resource] ?? [];
      groupedResources[resource] = [...existingEntries, scope];
      return groupedResources;
    },
    {}
  );
