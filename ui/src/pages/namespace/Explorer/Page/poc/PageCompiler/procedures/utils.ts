import { MutationType } from "../../schema/procedures/mutation";
import { QueryType } from "../../schema/procedures/query";

export const getUrl = (input: QueryType | MutationType) => {
  const { baseUrl, queryParams } = input;

  const searchParams = new URLSearchParams();
  queryParams?.forEach(({ key, value }) => {
    searchParams.append(key, value);
  });

  const queryString = searchParams.toString();
  return queryString ? baseUrl.concat("?", queryString) : baseUrl;
};
