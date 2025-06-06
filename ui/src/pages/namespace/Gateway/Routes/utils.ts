import {
  DirektivOpenApiSpecSchemaType,
  RouteMethod,
  routeMethods,
} from "src/api/gateway/schema";

function isRouteMethod(key: string): key is RouteMethod {
  return routeMethods.has(key as RouteMethod);
}

export const getMethodsFromOpenApiSpec = (
  spec: DirektivOpenApiSpecSchemaType
): RouteMethod[] => Object.keys(spec).filter(isRouteMethod);
