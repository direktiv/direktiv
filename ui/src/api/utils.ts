const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;

type AuthHeader =
  | {
      authorization: string;
    }
  | {
      "direktiv-token": string;
    };

export const getAuthHeader = (apiKey: string): AuthHeader => {
  if (isEnterprise) {
    return {
      authorization: `Bearer ${apiKey}`,
    };
  }

  return {
    "direktiv-token": apiKey,
  };
};

export const buildSearchParamsString = (
  searchParmsObj: Record<string, string | undefined>,
  withoutQuestionmark?: true
) => {
  const queryParams = new URLSearchParams();
  Object.entries(searchParmsObj).forEach(([name, value]) => {
    if (value) {
      queryParams.append(name, value);
    }
  });

  const queryParamsString = queryParams.toString();
  if (queryParamsString === "") {
    return queryParamsString;
  }

  return withoutQuestionmark ? queryParamsString : `?${queryParamsString}`;
};
