import * as React from "react";

import { ExtractQueryString, HandleError, apiKeyHeaders } from "../util";

import fetch from "isomorphic-fetch";

/*
    useRegistries is a react hook which returns create registry, delete registry and data
    takes:
      - url to direktiv api http://x/api/
      - namespace the namespace to query on
      - apikey to provide authentication of an apikey
*/

export const useDirektivRegistries = (url, namespace, apikey) => {
  const [data, setData] = React.useState(null);

  // getRegistries returns a list of registries
  const getRegistries = React.useCallback(
    async (...queryParameters) => {
      const resp = await fetch(
        `${url}functions/registries/namespaces/${namespace}${ExtractQueryString(
          false,
          ...queryParameters
        )}`,
        {
          headers: apiKeyHeaders(apikey),
        }
      );
      if (resp.ok) {
        const json = await resp.json();
        setData(json.registries);
        return await json.registries;
      } else {
        throw new Error(
          await HandleError("list registries", resp, "listRegistries")
        );
      }
    },
    [apikey, namespace, url]
  );

  React.useEffect(() => {
    const getData = async () => getRegistries();
    if (data === null) {
      getData().catch((e) => {
        console.error(e);
      });
    }
  }, [data, getRegistries]);

  async function createRegistry(key, val, ...queryParameters) {
    const resp = await fetch(
      `${url}functions/registries/namespaces/${namespace}${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        method: "POST",
        body: JSON.stringify({ data: val, reg: key }),
        headers: apiKeyHeaders(apikey),
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("create registry", resp, "createRegistry")
      );
    }
  }

  async function deleteRegistry(key, ...queryParameters) {
    const resp = await fetch(
      `${url}functions/registries/namespaces/${namespace}${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        method: "DELETE",
        body: JSON.stringify({
          reg: key,
        }),
        headers: apiKeyHeaders(apikey),
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("delete registry", resp, "deleteRegistry")
      );
    }
  }

  return {
    data,
    createRegistry,
    deleteRegistry,
    getRegistries,
  };
};
