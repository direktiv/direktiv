import * as React from "react";

import { ExtractQueryString, HandleError, apiKeyHeaders } from "../util";

import fetch from "isomorphic-fetch";

/*
    useBroadcastConfiguration is a react hook
    takes:
      - url to direktiv api http://x/api/
      - namespace the namespace to send the requests to
      - apikey to provide authentication of an apikey
*/
export const useDirektivBroadcastConfiguration = (url, namespace, apikey) => {
  const [data, setData] = React.useState(null);

  const getBroadcastConfiguration = React.useCallback(
    async (...queryParameters) => {
      const resp = await fetch(
        `${url}namespaces/${namespace}/config${ExtractQueryString(
          false,
          ...queryParameters
        )}`,
        {
          headers: apiKeyHeaders(apikey),
        }
      );
      if (!resp.ok) {
        throw new Error(
          await HandleError("fetch config", resp, "getNamespaceConfiguration")
        );
      }
      const json = await resp.json();
      setData(json);
      return json;
    },
    [apikey, namespace, url]
  );

  React.useEffect(() => {
    const getData = async () => getBroadcastConfiguration();
    if (data === null) {
      getData().catch((e) => {
        console.error(e);
      });
    }
  }, [data, getBroadcastConfiguration]);

  async function setBroadcastConfiguration(newconfig, ...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/config${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        method: "PATCH",
        body: newconfig,
        headers: apiKeyHeaders(apikey),
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("set config", resp, "setNamespaceConfiguration")
      );
    }
    return await resp.json();
  }

  return {
    data,
    getBroadcastConfiguration,
    setBroadcastConfiguration,
  };
};
