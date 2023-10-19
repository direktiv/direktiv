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
import fetch from "isomorphic-fetch";

/*
    useEvents is a react hook which returns details
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace the namespace to send the requests to
      - apikey to provide authentication of an apikey
*/
export const useDirektivEvents = (
  url,
  stream,
  namespace,
  apikey,
  queryParameters
) => {
  // DATA
  const [eventHistory, dispatchEventHistory] = React.useReducer(
    StateReducer,
    null
  );
  const [eventListeners, dispatchEventListeners] = React.useReducer(
    StateReducer,
    null
  );

  // ERRORS
  const [errHistory, setErrHistory] = React.useState(null);
  const [errListeners, setErrListeners] = React.useState(null);

  // Event history SSE
  const [eventHistorySource, setEventHistorySource] = React.useState(null);
  useEventSourceCleaner(eventHistorySource);

  // Event Listener SSE
  const [eventListenersSource, setEventListenersSource] = React.useState(null);
  useEventSourceCleaner(eventListenersSource);
  const [pathString, setPathString] = React.useState(null);

  // Store Query parameters
  const { queryString: eventHistoryQueryString } = useQueryString(
    false,
    queryParameters.history
  );
  const { queryString: eventListenersQueryString } = useQueryString(
    false,
    queryParameters.listeners
  );

  // Stores PageInfo about event list streams
  const [eventHistoryPageInfo, setEventHistoryPageInfo] = React.useState(null);
  const [eventListenersPageInfo, setEventListenersPageInfo] =
    React.useState(null);

  // Reset states when any prop that affects path is changed
  React.useEffect(() => {
    setEventHistoryPageInfo(null);
    setEventListenersPageInfo(null);
    dispatchEventHistory({ type: STATE.UPDATE, data: null });
    dispatchEventListeners({ type: STATE.UPDATE, data: null });
    setPathString(url && namespace ? `${url}namespaces/${namespace}` : null);
  }, [stream, namespace, url]);

  // Stream Event Source History Data Dispatch Handler
  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      if (json) {
        dispatchEventHistory({
          type: STATE.UPDATE,
          data: json.events.results,
        });

        setEventHistoryPageInfo(json.events.pageInfo);
      }
    }
    if (stream && pathString !== null) {
      // setup event listener
      const listener = new EventSourcePolyfill(
        `${pathString}/events${eventHistoryQueryString}`,
        {
          headers: apiKeyHeaders(apikey),
        }
      );

      listener.onerror = (e) => {
        genericEventSourceErrorHandler(e, setErrHistory);
      };

      listener.onmessage = (e) => readData(e);
      setEventHistorySource(listener);
    } else {
      setEventHistorySource(null);
    }
  }, [stream, apikey, eventHistoryQueryString, pathString]);

  // Stream Event Source Listeners Data Dispatch Handler
  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      if (json) {
        dispatchEventListeners({
          type: STATE.UPDATE,
          data: json.results,
        });

        setEventListenersPageInfo(json.pageInfo);
      }
    }
    if (stream && pathString !== null) {
      // setup event listener
      const listener = new EventSourcePolyfill(
        `${pathString}/event-listeners${eventListenersQueryString}`,
        {
          headers: apiKeyHeaders(apikey),
        }
      );

      listener.onerror = (e) => {
        genericEventSourceErrorHandler(e, setErrListeners);
      };

      listener.onmessage = (e) => readData(e);
      setEventListenersSource(listener);
    } else {
      setEventListenersSource(null);
    }
  }, [stream, apikey, eventListenersQueryString, pathString]);

  const getEventListeners = React.useCallback(
    async (...queryParameters) => {
      const resp = await fetch(
        `${url}namespaces/${namespace}/event-listeners${ExtractQueryString(
          false,
          ...queryParameters
        )}`,
        {
          method: "GET",
          headers: apiKeyHeaders(apikey),
        }
      );
      if (!resp.ok) {
        throw new Error(
          await HandleError("get event listeners", resp, "listEventHistory")
        );
      }
      return await resp.json();
    },
    [apikey, namespace, url]
  );

  const getEventHistory = React.useCallback(
    async (...queryParameters) => {
      const resp = await fetch(
        `${url}namespaces/${namespace}/events${ExtractQueryString(
          false,
          ...queryParameters
        )}`,
        {
          method: "GET",
          headers: apiKeyHeaders(apikey),
        }
      );
      if (!resp.ok) {
        throw new Error(
          await HandleError("get event history", resp, "listEventHistory")
        );
      }
      return await resp.json();
    },
    [apikey, namespace, url]
  );

  // Non Stream Data Dispatch Handler
  React.useEffect(() => {
    const update = async () => {
      if (!stream && pathString !== null && !errHistory && !errListeners) {
        setEventHistorySource(null);
        setEventListenersSource(null);

        const history = await getEventHistory();

        dispatchEventHistory({
          type: STATE.UPDATE,
          data: history.events.results,
        });

        setEventHistoryPageInfo(history.events.pageInfo);

        const listeners = await getEventListeners();
        dispatchEventListeners({
          type: STATE.UPDATE,
          data: listeners.results,
        });
        setEventListenersPageInfo(listeners.pageInfo);
      }
    };

    update();
  }, [
    stream,
    pathString,
    errHistory,
    errListeners,
    apikey,
    getEventHistory,
    getEventListeners,
  ]);

  async function replayEvent(event, ...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/events/${event}/replay${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        method: "POST",
        headers: {
          "content-type": "application/cloudevents+json; charset=UTF-8",
          ...apiKeyHeaders(apikey),
        },
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("send namespace event", resp, "sendNamespaceEvent")
      );
    }
    return;
  }

  async function sendEvent(event, ...queryParameters) {
    const resp = await fetch(
      `${url}namespaces/${namespace}/broadcast${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        method: "POST",
        body: event,
        headers: {
          "content-type": "application/cloudevents+json; charset=UTF-8",
          ...apiKeyHeaders(apikey),
        },
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("send namespace event", resp, "sendNamespaceEvent")
      );
    }
  }

  return {
    eventHistory,
    eventListeners,
    errHistory,
    errListeners,
    eventListenersPageInfo,
    eventHistoryPageInfo,
    getEventHistory,
    getEventListeners,
    sendEvent,
    replayEvent,
  };
};
