import * as React from "react";

import {
  ExtractQueryString,
  HandleError,
  STATE,
  StateReducer,
  apiKeyHeaders,
  genericEventSourceErrorHandler,
  useEventSourceCleaner,
  useQueryString,
} from "../util";

import { EventSourcePolyfill } from "event-source-polyfill";
// For testing
// import fetch from "cross-fetch"
// In Production
import fetch from "isomorphic-fetch";

/*
    useInstances is a react hook which returns a list of instances
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace the namespace to send the requests to
      - apikey to provide authentication of an apikey
*/
export const useDirektivInstances = (
  url,
  stream,
  namespace,
  apikey,
  ...queryParameters
) => {
  const [data, dispatchData] = React.useReducer(StateReducer, null);
  const [err, setErr] = React.useState(null);
  const [eventSource, setEventSource] = React.useState(null);
  useEventSourceCleaner(eventSource, "useInstances");

  // Store Query parameters
  const { queryString } = useQueryString(false, queryParameters);
  const [pathString, setPathString] = React.useState(null);

  // Stores PageInfo about instances list stream
  const [pageInfo, setPageInfo] = React.useState(null);

  // Stream Event Source Data Dispatch Handler
  React.useEffect(() => {
    const handler = setTimeout(() => {
      async function readData(e) {
        if (e.data === "") {
          return;
        }

        const json = JSON.parse(e.data);
        dispatchData({
          type: STATE.UPDATE,
          data: json.instances.results,
        });

        setPageInfo(json.instances.pageInfo);
      }
      if (stream && pathString !== null) {
        // setup event listener
        const listener = new EventSourcePolyfill(
          `${pathString}${queryString}`,
          {
            headers: apiKeyHeaders(apikey),
          }
        );

        listener.onerror = (e) => {
          genericEventSourceErrorHandler(e, setErr);
        };
        listener.onmessage = (e) => readData(e);
        setEventSource(listener);
      } else {
        setEventSource(null);
      }
    }, 50);

    return () => {
      clearTimeout(handler);
    };
  }, [stream, queryString, pathString, apikey]);

  // getInstances returns a list of instances
  const getInstances = React.useCallback(
    async (...queryParameters) => {
      // fetch instance list by default
      const resp = await fetch(
        `${url}namespaces/${namespace}/instances${ExtractQueryString(
          false,
          ...queryParameters
        )}`,
        {
          headers: apiKeyHeaders(apikey),
        }
      );
      if (!resp.ok) {
        throw new Error(
          await HandleError("list instances", resp, "listInstances")
        );
      }

      const json = await resp.json();
      setPageInfo(json.instances.pageInfo);
      return json.instances.results;
    },
    [apikey, namespace, url]
  );

  // Non Stream Data Dispatch Handler
  React.useEffect(() => {
    const update = async () => {
      if (!stream && pathString !== null && !err) {
        setEventSource(null);
        try {
          const instancesData = await getInstances();
          dispatchData({ type: STATE.UPDATE, data: instancesData });
        } catch (e) {
          setErr(e);
        }
      }
    };

    update();
  }, [stream, queryString, pathString, err, getInstances]);

  // Reset states when any prop that affects path is changed
  React.useEffect(() => {
    if (stream) {
      setPageInfo(null);
      setPathString(
        url && namespace ? `${url}namespaces/${namespace}/instances` : null
      );
    } else {
      dispatchData({ type: STATE.UPDATE, data: null });
      setPathString(
        url && namespace ? `${url}namespaces/${namespace}/instances` : null
      );
    }
  }, [stream, namespace, url]);

  return {
    data,
    err,
    pageInfo,
    getInstances,
  };
};
