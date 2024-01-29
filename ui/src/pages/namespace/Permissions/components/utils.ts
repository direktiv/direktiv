/**
 * takes a list of permissions and returns a list of scopes
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

export const permissionArrToScopes = (permissions: string[]) =>
  permissions.reduce((allScopes, permissionString) => {
    const scope = permissionString.split(":")?.[0];
    if (!scope) return allScopes;

    if (allScopes.includes(scope)) {
      return allScopes;
    }

    return [...allScopes, scope];
  }, [] as string[]);
