import * as React from "react";

import {
  ExtractQueryString,
  HandleError,
  STATE,
  SanitizePath,
  StateReducer,
  apiKeyHeaders,
  genericEventSourceErrorHandler,
  useEventSourceCleaner,
  useQueryString,
} from "../util";

import { EventSourcePolyfill } from "event-source-polyfill";
import { Templates } from "./templates";
import fetch from "isomorphic-fetch";

/*
  useNodes is a react hook which returns a list of items, createDirectory, createWorkflow, deleteDirectory, deleteWorkflow
  takes:
    - url to direktiv api http://x/api/
    - stream to use sse or a normal fetch
    - namespace the namespace to send the requests to
    - apikey to provide authentication of an apikey
*/
export const useDirektivNodes = (
  url,
  stream,
  namespace,
  path,
  apikey,
  ...queryParameters
) => {
  const [data, dispatchData] = React.useReducer(StateReducer, null);
  const [err, setErr] = React.useState(null);
  const [eventSource, setEventSource] = React.useState(null);
  useEventSourceCleaner(eventSource, "useNodes");

  // Store Query parameters
  const { queryString } = useQueryString(false, queryParameters);
  const [pathString, setPathString] = React.useState(null);

  // Stores PageInfo about node list stream
  const [pageInfo, setPageInfo] = React.useState(null);

  const templates = Templates;

  // Stream Event Source Data Dispatch Handler
  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }

      const json = JSON.parse(e.data);
      if (json?.children) {
        dispatchData({
          type: STATE.UPDATE,
          data: json,
        });

        setPageInfo(json.children.pageInfo);
      } else {
        dispatchData({
          type: STATE.UPDATE,
          data: json,
        });
      }
    }
    const handler = setTimeout(() => {
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

  const getNode = React.useCallback(async () => {
    let uri = `${url}namespaces/${namespace}/tree`;
    if (path !== "") {
      uri += `${SanitizePath(path)}`;
    }
    const resp = await fetch(`${uri}/${queryString}`, {
      headers: apiKeyHeaders(apikey),
    });
    if (resp.ok) {
      const json = await resp.json();
      if (json.children) {
        setPageInfo(json.children.pageInfo);
      }

      return json;
    } else {
      throw new Error(await HandleError("get node", resp, "listNodes"));
    }
  }, [apikey, namespace, path, queryString, url]);

  // Non Stream Data Dispatch Handler
  React.useEffect(() => {
    const update = async () => {
      if (!stream && pathString !== null && !err) {
        setEventSource(null);
        try {
          const nodeData = await getNode();
          dispatchData({ type: STATE.UPDATE, data: nodeData });
        } catch (e) {
          setErr(e);
        }
      }
    };
    update();
  }, [stream, queryString, pathString, err, getNode]);

  // Reset states when any prop that affects path is changed
  React.useEffect(() => {
    if (stream) {
      setPageInfo(null);
      setPathString(
        url && namespace && path
          ? `${url}namespaces/${namespace}/tree${SanitizePath(path)}`
          : null
      );
    } else {
      dispatchData({ type: STATE.UPDATE, data: null });
      setPathString(
        url && namespace && path
          ? `${url}namespaces/${namespace}/tree${SanitizePath(path)}`
          : null
      );
    }
  }, [stream, path, namespace, url]);

  async function createNode(name, type, yaml, ...queryParameters) {
    let uriPath = `${url}namespaces/${namespace}/tree`;
    if (path !== "") {
      uriPath += `${SanitizePath(path)}`;
    }
    const request = {
      method: "PUT",
      headers: apiKeyHeaders(apikey),
    };

    if (type === "workflow") {
      request.body = yaml;
      name += `?op=create-workflow`;
    } else {
      name += `?op=create-directory`;
    }
    const resp = await fetch(
      `${uriPath}/${name}${ExtractQueryString(true, ...queryParameters)}`,
      request
    );
    if (!resp.ok) {
      throw new Error(await HandleError("create node", resp));
    }

    return await resp.json();
  }

  async function createMirrorNode(name, mirrorSettings, ...queryParameters) {
    let uriPath = `${url}namespaces/${namespace}/tree`;
    if (path !== "") {
      uriPath += `${SanitizePath(path)}`;
    }
    const request = {
      method: "PUT",
      body: JSON.stringify(mirrorSettings),
      headers: apiKeyHeaders(apikey),
    };

    const resp = await fetch(
      `${uriPath}/${name}?op=create-directory${ExtractQueryString(
        true,
        ...queryParameters
      )}`,
      request
    );
    if (!resp.ok) {
      throw new Error(await HandleError("create node", resp));
    }

    return await resp.json();
  }

  async function deleteNode(name, recursive, ...queryParameters) {
    let uriPath = `${url}namespaces/${namespace}/tree`;
    if (path) {
      uriPath += `${SanitizePath(path)}`;
    }
    const resp = await fetch(
      `${uriPath}/${name}?op=delete-node&recursive=${
        recursive ? "true" : "false"
      }${ExtractQueryString(true, ...queryParameters)}`,
      {
        method: "DELETE",
        headers: apiKeyHeaders(apikey),
      }
    );
    if (!resp.ok) {
      throw new Error(await HandleError("delete node", resp, "deleteNode"));
    }
  }

  async function renameNode(fpath, oldname, newname, ...queryParameters) {
    let uriPath = `${url}namespaces/${namespace}/tree`;
    if (path) {
      uriPath += `${SanitizePath(fpath)}`;
    }
    const resp = await fetch(
      `${uriPath}${oldname}?op=rename-node${ExtractQueryString(
        true,
        ...queryParameters
      )}`,
      {
        method: "POST",
        body: JSON.stringify({ new: newname }),
        headers: apiKeyHeaders(apikey),
      }
    );
    if (!resp.ok) {
      throw new Error(await HandleError("rename node", resp, "renameNode"));
    }

    return await resp.json();
  }

  async function getWorkflowRouter(workflow, ...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/tree/${workflow}?op=router${ExtractQueryString(
        true,
        ...queryParameters
      )}`,
      {
        method: "get",
        headers: apiKeyHeaders(apikey),
      }
    );
    if (resp.ok) {
      const json = await resp.json();
      return json.live;
    } else {
      throw new Error(
        await HandleError("get workflow router", resp, "getWorkflow")
      );
    }
  }

  async function toggleWorkflow(workflow, active, ...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/tree/${workflow}?op=toggle${ExtractQueryString(
        true,
        ...queryParameters
      )}`,
      {
        method: "POST",
        body: JSON.stringify({
          live: active,
        }),
        headers: apiKeyHeaders(apikey),
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("toggle workflow", resp, "toggleWorkflow")
      );
    }

    return await resp.json();
  }

  return {
    data,
    err,
    templates,
    pageInfo,
    getNode,
    createNode,
    deleteNode,
    renameNode,
    toggleWorkflow,
    getWorkflowRouter,
    createMirrorNode,
  };
};
