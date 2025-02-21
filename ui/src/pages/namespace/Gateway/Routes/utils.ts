import {
  DirektivOpenApiSpecSchemaType,
  RouteMethod,
  routeMethods,
} from "src/api/gateway/schema";

export function isRouteMethod(key: string): key is RouteMethod {
  return routeMethods.has(key as RouteMethod);
}

export const getMethodFromOpenApiSpec = (
  spec: DirektivOpenApiSpecSchemaType
): RouteMethod[] => Object.keys(spec).filter(isRouteMethod);
