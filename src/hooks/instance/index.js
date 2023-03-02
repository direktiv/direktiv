import * as React from "react";

import {
  CloseEventSource,
  ExtractQueryString,
  HandleError,
  TrimPathSlashes,
  apiKeyHeaders,
} from "../util";

import { EventSourcePolyfill } from "event-source-polyfill";
import fetch from "isomorphic-fetch";

/*
    useInstanceLogs is a react hook which returns details for an instance
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace the namespace to send the requests to
      - instance the id used for the instance
      - apikey to provide authentication of an apikey
*/
export const useDirektivInstanceLogs = (
  url,
  stream,
  namespace,
  instance,
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

  // Stores PageInfo about instance log stream
  const [pageInfo, setPageInfo] = React.useState(null);

  // getInstanceLogs returns a list of logs
  const getInstanceLogs = React.useCallback(
    async (...queryParameters) => {
      // fetch instance list by default
      const resp = await fetch(
        `${url}namespaces/${namespace}/instances/${instance}/logs${ExtractQueryString(
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
      }

      throw new Error(
        await HandleError("get instance logs", resp, "instanceLogs")
      );
    },
    [apikey, instance, namespace, url]
  );

  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      if (Array.isArray(json.results)) {
        setData((old) => [...(old ?? []), ...json.results]);
        setPageInfo(json.pageInfo);
      }
    }

    if (stream) {
      if (eventSource.current === null) {
        // setup event listener
        const listener = new EventSourcePolyfill(
          `${url}namespaces/${namespace}/instances/${instance}/logs${queryString}`,
          {
            headers: apiKeyHeaders(apikey),
          }
        );

        listener.onerror = (e) => {
          if (e.status === 403) {
            setErr("permission denied");
          } else if (e.status === 404) {
            setErr(e.statusText);
          }
        };

        listener.onmessage = (e) => readData(e);
        eventSource.current = listener;
      }
    } else {
      if (data === null) {
        getInstanceLogs();
      }
    }
  }, [
    data,
    stream,
    apikey,
    url,
    namespace,
    instance,
    queryString,
    getInstanceLogs,
  ]);

  React.useEffect(() => {
    return () => {
      CloseEventSource(eventSource.current);
    };
  }, []);

  // If queryParameters change and streaming: update queryString, and reset sse connection
  React.useEffect(() => {
    if (stream) {
      const newQueryString = ExtractQueryString(false, ...queryParameters);
      if (newQueryString !== queryString) {
        setQueryString(newQueryString);
        CloseEventSource(eventSource.current);
        setData(null);
        eventSource.current = null;
      }
    }
  }, [queryParameters, queryString, stream]);

  return {
    data,
    err,
    pageInfo,
    getInstanceLogs,
  };
};
/*
    useInstance is a react hook which returns details for an instance
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace the namespace to send the requests to
      - instance the id used for the instance
      - apikey to provide authentication of an apikey
*/
export const useDirektivInstance = (
  url,
  stream,
  namespace,
  instance,
  apikey
) => {
  const [data, setData] = React.useState(null);
  const [latestRevision, setLatestRevision] = React.useState(null);
  const [workflow, setWorkflow] = React.useState(null);
  const [err, setErr] = React.useState(null);
  const [instanceID, setInstanceID] = React.useState(instance);
  const eventSource = React.useRef(null);

  const getLatestRevision = React.useCallback(
    async (workflowPath, ...queryParameters) => {
      // workflow doesnt exist anymore
      if (workflowPath === "") {
        setLatestRevision("");
      }

      const path = TrimPathSlashes(workflowPath);
      const resp = await fetch(
        `${url}namespaces/${namespace}/tree/${path}?op=validate-ref&ref=latest${ExtractQueryString(
          true,
          ...queryParameters
        )}`,
        {
          headers: apiKeyHeaders(apikey),
        }
      );
      if (resp.ok) {
        const json = await resp.json();
        setLatestRevision(json.revision.name);
        return json.revision.name;
      }
      throw new Error(
        await HandleError("get instance wf details", resp, "getInstance")
      );
    },
    [apikey, namespace, url]
  );

  // getInstance returns a list of instances
  const getInstance = React.useCallback(
    async (...queryParameters) => {
      // fetch instance list by default
      const resp = await fetch(
        `${url}namespaces/${namespace}/instances/${instanceID}${ExtractQueryString(
          false,
          ...queryParameters
        )}`,
        {
          headers: apiKeyHeaders(apikey),
        }
      );
      if (resp.ok) {
        const json = await resp.json();
        setData(json.instance);
        setWorkflow(json.workflow);
        getLatestRevision(json.workflow.path);
        return json.instance;
      }
      throw new Error(await HandleError("get instance", resp, "getInstance"));
    },
    [apikey, getLatestRevision, instanceID, namespace, url]
  );

  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      json.instance["flow"] = json.flow;
      setData(json.instance);
      setWorkflow(json.workflow);
      getLatestRevision(json.workflow.path);
    }
    if (stream) {
      if (eventSource.current === null) {
        // setup event listener
        const listener = new EventSourcePolyfill(
          `${url}namespaces/${namespace}/instances/${instanceID}`,
          {
            headers: apiKeyHeaders(apikey),
          }
        );

        listener.onerror = (e) => {
          if (e.status === 403) {
            setErr("permission denied");
          } else if (e.status === 404) {
            setErr(e.statusText);
          } else {
            try {
              const json = JSON.parse(e.data);
              setErr(json.Message);
            } catch (e) {
              console.error(e);
            }
          }
        };

        listener.onmessage = (e) => readData(e);
        eventSource.current = listener;
      }
    } else {
      if (data === null) {
        getInstance();
      }
    }
  }, [
    apikey,
    data,
    getInstance,
    getLatestRevision,
    instanceID,
    namespace,
    stream,
    url,
  ]);

  // If instance changes reset eventSource

  // TODO: check if this is really necessary
  React.useEffect(() => {
    if (stream) {
      if (instance !== instanceID) {
        setInstanceID(instance);
        CloseEventSource(eventSource.current);
        eventSource.current = null;
        setData(null);
      }
    }
  }, [instanceID, instance, stream]);

  React.useEffect(() => {
    return () => {
      CloseEventSource(eventSource.current);
    };
  }, []);

  async function getInput(...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/instances/${instanceID}/input${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        method: "GET",
        headers: apiKeyHeaders(apikey),
      }
    );
    if (resp.ok) {
      const json = await resp.json();
      return atob(json.data);
    }
    throw new Error(
      await HandleError("get instance input", resp, "getInstance")
    );
  }

  async function getOutput(...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/instances/${instanceID}/output${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        method: "GET",
        headers: apiKeyHeaders(apikey),
      }
    );
    if (resp.ok) {
      const json = await resp.json();
      return atob(json.data);
    }
    throw new Error(
      await HandleError("get instance output", resp, "getInstance")
    );
  }

  async function cancelInstance(...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/instances/${instanceID}/cancel${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        method: "POST",
        headers: apiKeyHeaders(apikey),
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("cancelling instance", resp, "cancelInstance")
      );
    }
  }

  return {
    data,
    workflow,
    latestRevision,
    err,
    getInstance,
    cancelInstance,
    getInput,
    getOutput,
  };
};
