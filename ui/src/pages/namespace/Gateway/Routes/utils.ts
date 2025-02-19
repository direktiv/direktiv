import { MethodsObject, routeMethods } from "src/api/gateway/schema";

// Helper functions to extract the methods from the spec object
export function isRouteMethod(
  key: string
): key is (typeof routeMethods)[number] {
  return routeMethods.includes(key as (typeof routeMethods)[number]);
}

export function getMethodOperations(
  spec: Record<string, unknown>
): MethodsObject {
  return Object.entries(spec)
    .filter(([key]) => isRouteMethod(key))
    .reduce((acc, [key, value]) => {
      acc[key as (typeof routeMethods)[number]] = value;
      return acc;
    }, {} as MethodsObject);
}
