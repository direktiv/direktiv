const splitPermissionString = (permission: string): [string, string] => {
  const [scope, resource] = permission.split(":");
  return [scope ?? "", resource ?? ""];
};

/**
 * takes a list of permissions and returns all availables of scopes
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
export const permissionsToScopes = (permissions: string[]) =>
  permissions.reduce((allScopes, permissionString) => {
    const [scope] = splitPermissionString(permissionString);
    if (allScopes.includes(scope)) {
      return allScopes;
    }

    return [...allScopes, scope];
  }, [] as string[]);
