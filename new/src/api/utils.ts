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
