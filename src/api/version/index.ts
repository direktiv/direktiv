import { CommonApiParams } from "../types";

export const getVersion = async ({
  apiKey,
}: CommonApiParams): Promise<GatewayListItem[]> => {
  const res = await fetch(`/api/version`, {
    method: "GET",
    headers: {
      "direktiv-token": `${apiKey}`,
    },
  });

  // return gatewayMocks;

  if (res.ok) {
    try {
      return (await res.json()) as Promise<GatewayListItem[]>;
    } catch (error) {
      return Promise.reject("could not format response");
    }
  }
  return Promise.reject(`${res.status}`);
};
