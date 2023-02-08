import { ExtractQueryString, HandleError, apiKeyHeaders } from "../util";

import fetch from "cross-fetch";

/*
    useNamespaceMetrics is a react hook which metric details
    takes:
      - url to direktiv api http://x/api/
      - namespace to use with the api
      - apikey to provide authentication of an apikey
*/
export const useDirektivNamespaceMetrics = (url, namespace, apikey) => {
  async function getInvoked(...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/metrics/invoked${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        headers: apiKeyHeaders(apikey),
      }
    );
    if (resp.ok) {
      return await resp.json();
    } else {
      throw new Error(
        await HandleError("get invoked metrics", resp, "getMetrics")
      );
    }
  }

  async function getSuccessful(...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/metrics/successful${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        headers: apiKeyHeaders(apikey),
      }
    );
    if (resp.ok) {
      return await resp.json();
    } else {
      throw new Error(
        await HandleError("get successful metrics", resp, "getMetrics")
      );
    }
  }

  async function getFailed(...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/metrics/failed${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        headers: apiKeyHeaders(apikey),
      }
    );
    if (resp.ok) {
      return await resp.json();
    } else {
      throw new Error(
        await HandleError("get failed metrics", resp, "getMetrics")
      );
    }
  }

  async function getMilliseconds(...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/metrics/milliseconds${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        headers: apiKeyHeaders(apikey),
      }
    );
    if (resp.ok) {
      return await resp.json();
    } else {
      throw new Error(
        await HandleError("get millisecond metrics", resp, "getMetrics")
      );
    }
  }

  return {
    getInvoked,
    getSuccessful,
    getFailed,
    getMilliseconds,
  };
};
