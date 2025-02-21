import {
  MethodsObject,
  RouteMethod,
  routeMethods,
} from "src/api/gateway/schema";

// Helper functions to extract the methods from the spec object
export function isRouteMethod(key: string): key is RouteMethod {
  return routeMethods.has(key as RouteMethod);
}

export function getMethodOperations(
  spec: Record<string, unknown>
): MethodsObject {
  return Object.entries(spec)
    .filter(([key]) => isRouteMethod(key))
    .reduce((acc, [key, value]) => {
      acc[key as RouteMethod] = value as Record<string, unknown>;
      return acc;
    }, {} as MethodsObject);
}
