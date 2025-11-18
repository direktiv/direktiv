import { DirektivPagesType } from "../schema";
import { QueryType } from "../schema/procedures/query";
import { keyValueArrayToObject } from "../PageCompiler/primitives/keyValue/utils";

export const clonePage = (page: DirektivPagesType): DirektivPagesType =>
  structuredClone(page);

export const queryToUrl = (query: QueryType) => {
  let { url } = query;

  const searchParams = new URLSearchParams(
    keyValueArrayToObject(query.queryParams ?? [])
  );

  const paramsString = searchParams.toString();

  if (paramsString) {
    url = url.concat("?", paramsString);
  }

  return url;
};
