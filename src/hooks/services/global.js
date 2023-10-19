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
    useGlobalServiceRevision
    - url
    - service
    - revision
    - apikey
*/
export const useDirektivGlobalServiceRevision = (
  url,
  service,
  revision,
  apikey
) => {
  const [revisionDetails, setRevisionDetails] = React.useState(null);
  const [pods, setPods] = React.useState([]);
  const [err, setErr] = React.useState(null);
  const podSource = React.useRef(null);
  const revisionSource = React.useRef(null);
  const podsRef = React.useRef(pods);

  React.useEffect(() => {
    async function readData(e) {
      const podz = podsRef.current;

      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);

      switch (json.event) {
        case "DELETED":
          for (var i = 0; i < pods.length; i++) {
            if (podz[i].name === json.pod.name) {
              podz.splice(i, 1);
              podsRef.current = pods;
              break;
            }
          }
          break;
        case "MODIFIED":
          for (i = 0; i < podz.length; i++) {
            if (podz[i].name === json.pod.name) {
              podz[i] = json.pod;
              podsRef.current = podz;
              break;
            }
          }
          break;
        default: {
          let found = false;
          for (i = 0; i < podz.length; i++) {
            if (podz[i].name === json.pod.name) {
              found = true;
              break;
            }
          }
          if (!found) {
            podz.push(json.pod);
            podsRef.current = pods;
          }
        }
      }
      setPods(JSON.parse(JSON.stringify(podsRef.current)));
    }

    if (podSource.current === null) {
      const listener = new EventSourcePolyfill(
        `${url}functions/${service}/revisions/${revision}/pods`,
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
      podSource.current = listener;
    }
  }, [apikey, pods, revision, service, url]);

  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      if (json.event === "ADDED" || json.event === "MODIFIED") {
        setRevisionDetails(json.revision);
      }
      // if (json.event === "DELETED") {
      //     history.goBack()
      // }
    }

    if (revisionSource.current === null) {
      // setup event listener
      const listener = new EventSourcePolyfill(
        `${url}functions/${service}/revisions/${revision}`,
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
      revisionSource.current = listener;
    }
  }, [apikey, url, service, revision]);

  React.useEffect(() => {
    return () => {
      CloseEventSource(revisionSource.current);
      CloseEventSource(podSource.current);
    };
  }, []);

  return {
    revisionDetails,
    pods,
    err,
  };
};

/* 
    useGlobalService takes
    - url
    - service
    - navigate (react router navigation to go backwards if revision is deleted)
    - apikey
*/
export const useDirektivGlobalService = (url, service, navigate, apikey) => {
  const [revisions, setRevisions] = React.useState(null);
  const [fn, setFn] = React.useState(null);
  const [traffic, setTraffic] = React.useState(null);
  const [config, setConfig] = React.useState(null);
  const revisionsRef = React.useRef(revisions ? revisions : []);

  const [err, setErr] = React.useState(null);
  const trafficSource = React.useRef(null);
  const eventSource = React.useRef(null);

  React.useEffect(() => {
    async function readData(e) {
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);

      if (json.event === "MODIFIED" || json.event === "ADDED") {
        setFn(JSON.parse(JSON.stringify(json.function)));
        setTraffic(JSON.parse(JSON.stringify(json.traffic)));
      }
    }
    if (trafficSource.current === null) {
      // setup event listener
      const listener = new EventSourcePolyfill(`${url}functions/${service}`, {
        headers: apiKeyHeaders(apikey),
      });

      listener.onerror = (e) => {
        if (e.status === 404) {
          setErr(e.statusText);
        } else if (e.status === 403) {
          setErr("permission denied");
        }
      };

      listener.onmessage = (e) => readData(e);
      trafficSource.current = listener;
    }
  }, [fn, apikey, url, service]);

  React.useEffect(() => {
    async function readData(e) {
      const revs = revisionsRef.current;
      if (e.data === "") {
        return;
      }
      const json = JSON.parse(e.data);
      switch (json.event) {
        case "DELETED":
          for (var i = 0; i < revs.length; i++) {
            if (revs[i].name === json.revision.name) {
              revs.splice(i, 1);
              revisionsRef.current = revs;
              break;
            }
          }
          if (revs.length === 0) {
            navigate(-1);
          }
          break;
        case "MODIFIED":
          for (i = 0; i < revs.length; i++) {
            if (revs[i].name === json.revision.name) {
              revs[i] = json.revision;
              revisionsRef.current = revs;
              break;
            }
          }
          break;
        default: {
          let found = false;
          for (i = 0; i < revs.length; i++) {
            if (revs[i].name === json.revision.name) {
              found = true;
              break;
            }
          }
          if (!found) {
            revs.push(json.revision);
            revisionsRef.current = revs;
          }
        }
      }

      setRevisions(JSON.parse(JSON.stringify(revisionsRef.current)));
    }

    if (eventSource.current === null) {
      // setup event listener
      const listener = new EventSourcePolyfill(
        `${url}functions/${service}/revisions`,
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
  }, [revisions, apikey, url, service, navigate]);

  React.useEffect(() => {
    return () => {
      CloseEventSource(eventSource.current);
      CloseEventSource(trafficSource.current);
    };
  }, []);

  async function createGlobalServiceRevision(
    image,
    minScale,
    size,
    cmd,
    traffic,
    ...queryParameters
  ) {
    const resp = await fetch(
      `${url}functions/${service}${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        headers: apiKeyHeaders(apikey),
        method: "POST",
        body: JSON.stringify({
          trafficPercent: traffic,
          cmd,
          image,
          minScale,
          size,
        }),
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError(
          "create global service revision",
          resp,
          "createRevision"
        )
      );
    }
  }

  async function deleteGlobalServiceRevision(rev, ...queryParameters) {
    const resp = await fetch(
      `${url}functions/${service}/revisions/${rev}${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        headers: apiKeyHeaders(apikey),
        method: "DELETE",
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError(
          "delete global service revision",
          resp,
          "deleteRevision"
        )
      );
    }
  }

  async function getServiceConfig(...queryParameters) {
    const resp = await fetch(
      `${url}functions/${service}${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        headers: apiKeyHeaders(apikey),
        method: "GET",
      }
    );
    if (resp.ok) {
      const json = await resp.json();
      setConfig(json.config);
      return json.config;
    } else {
      throw new Error(
        await HandleError("get global service", resp, "getService")
      );
    }
  }

  async function setGlobalServiceRevisionTraffic(
    rev1,
    rev1value,
    rev2,
    rev2value,
    ...queryParameters
  ) {
    const resp = await fetch(
      `${url}functions/${service}${ExtractQueryString(
        false,
        ...queryParameters
      )}`,
      {
        method: "PATCH",
        headers: apiKeyHeaders(apikey),
        body: JSON.stringify({
          values: [
            {
              revision: rev1,
              percent: rev1value,
            },
            {
              revision: rev2,
              percent: rev2value,
            },
          ],
        }),
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError(
          "update traffic global service",
          resp,
          "updateTraffic"
        )
      );
    }
  }

  return {
    revisions,
    fn,
    config,
    traffic,
    err,
    createGlobalServiceRevision,
    getServiceConfig,
    deleteGlobalServiceRevision,
    setGlobalServiceRevisionTraffic,
  };
};

/*
    useGlobalServices is a react hook 
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - navigate (react-router object to navigate backwards)
      - apikey to provide authentication of an apikey
*/
export const useDirektivGlobalServices = (url, stream, apikey) => {
  const [data, setData] = React.useState(null);
  const [config, setConfig] = React.useState(null);

  const functionsRef = React.useRef(data ? data : []);
  const [err, setErr] = React.useState(null);
  const eventSource = React.useRef(null);

  const getGlobalServices = React.useCallback(
    async (...queryParameters) => {
      const resp = await fetch(
        `${url}functions${ExtractQueryString(false, ...queryParameters)}`,
        {
          headers: apiKeyHeaders(apikey),
          method: "GET",
        }
      );
      if (resp.ok) {
        const json = await resp.json();
        setData(json.functions);
        return json.functions;
      } else {
        throw new Error(
          await HandleError("get global services", resp, "listServices")
        );
      }
    },
    [apikey, url]
  );

  React.useEffect(() => {
    async function readData(e) {
      const funcs = functionsRef.current;
      if (e.data === "") {
        if (funcs === null) {
          setData([]);
        }
        return;
      }
      const json = JSON.parse(e.data);
      switch (json.event) {
        case "DELETED":
          for (var i = 0; i < funcs.length; i++) {
            if (funcs[i].serviceName === json.function.serviceName) {
              funcs.splice(i, 1);
              functionsRef.current = funcs;
              break;
            }
          }
          break;
        case "MODIFIED":
          for (i = 0; i < funcs.length; i++) {
            if (funcs[i].serviceName === json.function.serviceName) {
              funcs[i] = json.function;
              functionsRef.current = funcs;
              break;
            }
          }
          break;
        default: {
          let found = false;
          for (i = 0; i < funcs.length; i++) {
            if (funcs[i].serviceName === json.function.serviceName) {
              found = true;
              break;
            }
          }
          if (!found) {
            funcs.push(json.function);
            functionsRef.current = funcs;
          }
        }
      }
      setData(JSON.parse(JSON.stringify(functionsRef.current)));
    }
    if (stream) {
      if (eventSource.current === null) {
        // setup event listener
        const listener = new EventSourcePolyfill(`${url}functions`, {
          headers: apiKeyHeaders(apikey),
        });

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
        getGlobalServices();
      }
    }
  }, [data, apikey, stream, url, getGlobalServices]);

  React.useEffect(() => {
    return () => {
      CloseEventSource(eventSource.current);
    };
  }, []);

  async function getConfig(...queryParameters) {
    const resp = await fetch(
      `${url}functions${ExtractQueryString(false, ...queryParameters)}`,
      {
        headers: apiKeyHeaders(apikey),
        method: "GET",
      }
    );
    if (resp.ok) {
      const json = await resp.json();
      setConfig(json.config);
      return json.config;
    } else {
      throw new Error(
        await HandleError("get namespace service", resp, "listServices")
      );
    }
  }

  async function createGlobalService(
    name,
    image,
    minScale,
    size,
    cmd,
    ...queryParameters
  ) {
    const resp = await fetch(
      `${url}functions${ExtractQueryString(false, ...queryParameters)}`,
      {
        headers: apiKeyHeaders(apikey),
        method: "POST",
        body: JSON.stringify({
          cmd,
          image,
          minScale,
          name,
          size,
        }),
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("create global service", resp, "createService")
      );
    }
  }

  async function deleteGlobalService(name, ...queryParameters) {
    const resp = await fetch(
      `${url}functions/${name}${ExtractQueryString(false, ...queryParameters)}`,
      {
        headers: apiKeyHeaders(apikey),
        method: "DELETE",
      }
    );
    if (!resp.ok) {
      throw new Error(
        await HandleError("delete global service", resp, "deleteService")
      );
    }
  }

  return {
    data,
    err,
    config,
    getGlobalServices,
    getConfig,
    createGlobalService,
    deleteGlobalService,
  };
};
