import * as React from "react";

import {
  HandleError,
  STATE,
  StateReducer,
  apiKeyHeaders,
  genericEventSourceErrorHandler,
  useEventSourceCleaner,
  useQueryString,
} from "../util";

import { EventSourcePolyfill } from "event-source-polyfill";
import fetch from "isomorphic-fetch";

export const useDirektivMirrorLogs = (
  url,
  stream,
  namespace,
  activity,
  apikey,
  ...queryParameters
) => {
  const [data, dataDispatch] = React.useReducer(StateReducer, null);
  const [err, setErr] = React.useState(null);

  const [eventSource, setEventSource] = React.useState(null);
  useEventSourceCleaner(eventSource);

  const { queryString } = useQueryString(false, queryParameters);
  const [pathString, setPathString] = React.useState(null);

  // Stores PageInfo about node list stream
  const [pageInfo, setPageInfo] = React.useState(null);

  // Stream Event Source Data Dispatch Handler
  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      if (json) {
        dataDispatch({
          type: STATE.APPENDLIST,
          data: json.results,
        });

        setPageInfo(json.pageInfo);
      }
    }
    if (stream && pathString !== null) {
      // setup event listener
      const listener = new EventSourcePolyfill(`${pathString}${queryString}`, {
        headers: apiKeyHeaders(apikey),
      });

      listener.onerror = (e) => {
        genericEventSourceErrorHandler(e, setErr);
      };

      listener.onmessage = (e) => readData(e);
      setEventSource(listener);
    } else {
      setEventSource(null);
    }
  }, [stream, queryString, pathString, apikey]);

  const getActivityLogs = React.useCallback(async () => {
    const request = {
      method: "GET",
      headers: apiKeyHeaders(apikey),
    };

    const resp = await fetch(`${pathString}${queryString}`, request);
    if (!resp.ok) {
      throw new Error(
        await HandleError("mirror activity logs", resp, "mirrorActivityLogs")
      );
    }

    return await resp.json();
  }, [apikey, pathString, queryString]);

  // Non Stream Data Dispatch Handler
  React.useEffect(() => {
    if (!stream && pathString !== null) {
      setEventSource(null);
      getActivityLogs().then((data) => {
        dataDispatch({ type: STATE.UPDATE, data: data });
      });
    }
  }, [stream, queryString, pathString, getActivityLogs]);

  // Reset states when any prop that affects path is changed
  React.useEffect(() => {
    if (stream) {
      setPageInfo(null);
      dataDispatch({ type: STATE.UPDATE, data: null });
      setPathString(
        url && namespace && activity
          ? `${url}namespaces/${namespace}/activities/${activity}/logs`
          : null
      );
    } else {
      dataDispatch({ type: STATE.UPDATE, data: null });
      setPathString(
        url && namespace && activity
          ? `${url}namespaces/${namespace}/activities/${activity}/logs`
          : null
      );
    }
  }, [stream, activity, namespace, url]);

  return {
    data,
    err,
    pageInfo,
    getActivityLogs,
  };
};
