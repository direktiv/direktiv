import { isEnterprise } from "~/config/env/utils";

type AuthHeader =
  | {
      authorization: string;
    }
  | {
      "direktiv-api-key": string;
    };

export const getAuthHeader = (apiKey: string): AuthHeader => {
  if (isEnterprise()) {
    return {
      authorization: `Bearer ${apiKey}`,
    };
  }

  return {
    "direktiv-api-key": apiKey,
  };
};

const convertDateToISOString = (value: string | number | Date) => {
  const isDate = value instanceof Date;
  return isDate ? value.toISOString() : value;
};

export const buildSearchParamsString = (
  searchParmsObj: Record<string, string | Date | number | undefined>,
  withoutQuestionmark?: true
) => {
  const queryParams = new URLSearchParams();
  Object.entries(searchParmsObj).forEach(([name, value]) => {
    if (value) {
      const queryValue = convertDateToISOString(value);
      queryParams.append(name, `${queryValue}`);
    }
  });

  const queryParamsString = queryParams.toString();
  if (queryParamsString === "") {
    return queryParamsString;
  }

  return withoutQuestionmark ? queryParamsString : `?${queryParamsString}`;
};
