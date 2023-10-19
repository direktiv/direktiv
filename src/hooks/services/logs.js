import * as React from "react";

import { CloseEventSource, apiKeyHeaders } from "../util";

import { EventSourcePolyfill } from "event-source-polyfill";

/*
    usePodLogs
    - url
    - pod
    - apikey
*/
export const useDirektivPodLogs = (url, pod, apikey) => {
  const [data, setData] = React.useState(null);
  const [err, setErr] = React.useState(null);
  const eventSource = React.useRef(null);

  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      setData(json);
    }

    if (eventSource.current === null) {
      // setup event listener
      const listener = new EventSourcePolyfill(
        `${url}functions/logs/pod/${pod}`,
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
  }, [apikey, pod, url]);

  React.useEffect(() => {
    return () => {
      // cleanup
      CloseEventSource(eventSource.current);
    };
  }, []);

  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      setData(json);
    }

    if (eventSource.current !== null) {
      // setup event listener
      const listener = new EventSourcePolyfill(
        `${url}functions/logs/pod/${pod}`,
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
  }, [pod, apikey, url]);

  return {
    data,
    err,
  };
};
