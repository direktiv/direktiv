import * as React from "react";

import {
  CloseEventSource,
  ExtractQueryString,
  HandleError,
  apiKeyHeaders,
} from "../util";

import { EventSourcePolyfill } from "event-source-polyfill";
import fetch from "isomorphic-fetch";

/*
    useNamespaceLogs is a react hook which returns data, err or getNamespaceLogs()
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace to call the api on
      - apikey to provide authentication of an apikey
*/
export const useDirektivNamespaceLogs = (
  url,
  stream,
  namespace,
  apikey,
  ...queryParameters
) => {
  const [data, setData] = React.useState(null);
  const [err, setErr] = React.useState(null);
  const eventSource = React.useRef(null);

  // Store Query parameters
  const [queryString, setQueryString] = React.useState(
    ExtractQueryString(false, ...queryParameters)
  );

  // Stores PageInfo about namespace log stream
  const [pageInfo, setPageInfo] = React.useState(null);

  // getNamespaces returns a list of namespaces
  const getNamespaceLogs = React.useCallback(
    async (...queryParameters) => {
      // fetch namespace list by default
      const resp = await fetch(
        `${url}namespaces/${namespace}/logs${ExtractQueryString(
          false,
          ...queryParameters
        )}`,
        {
          headers: apiKeyHeaders(apikey),
        }
      );
      if (resp.ok) {
        const json = await resp.json();
        setData(json.results);
        setPageInfo(json.pageInfo);
        return json.results;
      } else {
        throw new Error(
          await HandleError("list namespace logs", resp, "namespaceLogs")
        );
      }
    },
    [apikey, namespace, url]
  );

  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      setData(json.results);
      setPageInfo(json.pageInfo);
    }
    if (stream) {
      if (eventSource.current === null) {
        // setup event listener
        const listener = new EventSourcePolyfill(
          `${url}namespaces/${namespace}/logs${queryString}`,
          {
            headers: apiKeyHeaders(apikey),
          }
        );

        listener.onerror = (e) => {
          if (e.status === 404) {
            setErr(e.statusText);
          } else if (e.status === 403) {
            setErr("permission denied");
          }
        };
        listener.onmessage = (e) => readData(e);
        eventSource.current = listener;
      }
    } else {
      if (data === null) {
        getNamespaceLogs();
      }
    }
  }, [data, apikey, stream, url, namespace, queryString, getNamespaceLogs]);

  // If queryParameters change and streaming: update queryString, and reset sse connection
  React.useEffect(() => {
    if (stream) {
      const newQueryString = ExtractQueryString(false, ...queryParameters);
      if (newQueryString !== queryString) {
        setQueryString(newQueryString);
        CloseEventSource(eventSource.current);
        eventSource.current = null;
      }
    }
  }, [queryParameters, queryString, stream]);

  React.useEffect(() => {
    return () => {
      CloseEventSource(eventSource.current);
    };
  }, []);

  return {
    data,
    err,
    pageInfo,
    getNamespaceLogs,
  };
};
