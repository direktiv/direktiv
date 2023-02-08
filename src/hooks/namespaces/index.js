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
    useNamespaces is a react hook which returns createNamespace, deleteNamespace and data
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - apikey to provide authentication of an apikey
*/
export const useDirektivNamespaces = (
  url,
  stream,
  apikey,
  ...queryParameters
) => {
  const [data, setData] = React.useState(null);
  const [load, setLoad] = React.useState(true);
  const [err, setErr] = React.useState(null);
  const eventSource = React.useRef(null);

  // Store Query parameters
  const [queryString, setQueryString] = React.useState(
    ExtractQueryString(false, ...queryParameters)
  );

  // Stores PageInfo about namespace list stream
  const [pageInfo, setPageInfo] = React.useState(null);

  // getNamespaces returns a list of namespaces and update the internal state
  const getNamespaces = React.useCallback(
    async (...queryParameters) => {
      // fetch namespace list by default
      const resp = await fetch(
        `${url}namespaces${ExtractQueryString(false, ...queryParameters)}`,
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
          await HandleError("list namespaces", resp, "listNamespaces")
        );
      }
    },
    [apikey, url]
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
          `${url}namespaces${queryString}`,
          {
            headers: apiKeyHeaders(apikey),
          }
        );

        listener.onerror = (e) => {
          setErr(e);
          if (e.status === 404) {
            setErr(e.statusText);
          } else if (e.status === 403) {
            setErr("permission denied");
          }
        };

        listener.onmessage = (e) => readData(e);
        eventSource.current = listener;
        setLoad(false);
        setErr("");
      }
    } else {
      if (data === null) {
        getNamespaces();
      }
    }
  }, [apikey, data, getNamespaces, queryString, stream, url]);

  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      setData(json.results);
      setPageInfo(json.pageInfo);
    }
    if (!load && eventSource.current !== null) {
      CloseEventSource(eventSource.current);
      // setup event listener
      const listener = new EventSourcePolyfill(
        `${url}namespaces${queryString}`,
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
      setErr("");
    }
  }, [apikey, load, queryString, url]);

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

  // createNamespace creates a namespace from direktiv
  async function createNamespace(namespace, ...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        method: "PUT",
        headers: apiKeyHeaders(apikey),
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("create a namespace", resp, "addNamespace")
      );
    }
  }

  async function createMirrorNamespace(
    namespace,
    mirrorSettings,
    ...queryParameters
  ) {
    const request = {
      method: "PUT",
      body: JSON.stringify(mirrorSettings),
      headers: apiKeyHeaders(apikey),
    };

    const resp = await fetch(
      `${url}namespaces/${namespace}${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      request
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("create a mirror namespace", resp, "addNamespace")
      );
    }
  }

  // deleteNamespace deletes a namespace from direktiv
  async function deleteNamespace(namespace, ...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}?recursive=true${ExtractQueryString(
        true,
        ...queryParameters
      )}`,
      {
        method: "DELETE",
        headers: apiKeyHeaders(apikey),
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("delete a namespace", resp, "deleteNamespace")
      );
    }
  }

  return {
    data,
    err,
    pageInfo,
    createNamespace,
    deleteNamespace,
    getNamespaces,
    createMirrorNamespace,
  };
};
