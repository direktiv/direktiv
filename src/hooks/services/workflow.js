import * as React from 'react'
import { EventStateReducer } from '../util'
import { EVENTSTATE } from '../util'
import { HandleError, ExtractQueryString, SanitizePath, StateReducer, STATE, useEventSourceCleaner, useQueryString, genericEventSourceErrorHandler, CloseEventSource, apiKeyHeaders } from '../util'
const { EventSourcePolyfill } = require('event-source-polyfill')
const fetch = require('isomorphic-fetch')
/* 
    useNamespaceServiceRevision takes
    - url
    - namespace
    - path
    - service
    - version
    - revision
    - apikey
*/
export const useDirektivWorkflowServiceRevision = (url, namespace, path, service, version, revision, apikey) => {
    const [revisionDetails, setRevisionDetails] = React.useState(null)
    const [podSource, setPodSource] = React.useState(null)
    const [pods, setPods] = React.useState([])
    const [err, setErr] = React.useState(null)
    const [revisionSource, setRevisionSource] = React.useState(null)

    const podsRef = React.useRef(pods)


    React.useEffect(() => {
        if (podSource === null) {
            let listener = new EventSourcePolyfill(`${url}functions/namespaces/${namespace}/tree/${path}?op=pods&svn=${service}&rev=${revision}&version=${version}`, {
                headers: apiKeyHeaders(apikey)
            })

            listener.onerror = (e) => {
                if (e.status === 404) {
                    setErr(e.statusText)
                } else if (e.status === 403) {
                    setErr("permission denied")
                }
            }

            async function readData(e) {
                let podz = podsRef.current

                if (e.data === "") {
                    return
                }
                let json = JSON.parse(e.data)

                switch (json.event) {
                    case "DELETED":
                        for (var i = 0; i < pods.length; i++) {
                            if (podz[i].name === json.pod.name) {
                                podz.splice(i, 1)
                                podsRef.current = pods
                                break
                            }
                        }
                        break
                    case "MODIFIED":
                        for (i = 0; i < podz.length; i++) {
                            if (podz[i].name === json.pod.name) {
                                podz[i] = json.pod
                                podsRef.current = podz
                                break
                            }
                        }
                        break
                    default:
                        let found = false
                        for (i = 0; i < podz.length; i++) {
                            if (podz[i].name === json.pod.name) {
                                found = true
                                break
                            }
                        }
                        if (!found) {
                            podz.push(json.pod)
                            podsRef.current = pods
                        }
                }
                setPods(JSON.parse(JSON.stringify(podsRef.current)))

            }
            listener.onmessage = e => readData(e)
            setPodSource(listener)
        }
    })

    React.useEffect(() => {
        if (revisionSource === null) {
            // setup event listener 
            let listener = new EventSourcePolyfill(`${url}functions/namespaces/${namespace}/tree/${path}?op=function-revision&svn=${service}&rev=${revision}&version=${version}`, {
                headers: apiKeyHeaders(apikey)
            })

            listener.onerror = (e) => {
                if (e.status === 404) {
                    setErr(e.statusText)
                } else if (e.status === 403) {
                    setErr("permission denied")
                }
            }

            async function readData(e) {
                if (e.data === "") {
                    return
                }
                let json = JSON.parse(e.data)
                if (json.event === "ADDED" || json.event === "MODIFIED") {
                    setRevisionDetails(json.revision)
                }
                // if (json.event === "DELETED") {
                //     history.goBack()
                // }
            }

            listener.onmessage = e => readData(e)
            setRevisionSource(listener)
        }
    }, [revisionSource])

    React.useEffect(() => {
        return () => {
            CloseEventSource(revisionSource)
            CloseEventSource(podSource)
        }
    }, [revisionSource, podSource])


    return {
        revisionDetails,
        pods,
        err
    }
}

/* 
    useWorkflowService
    - url
    - namespace
    - path
    - service
    - version
    - navigate(react router object to navigate backwards)
    - apikey
*/
export const useDirektivWorkflowService = (url, namespace, path, service, version, navigate, apikey) => {
    const [revisions, setRevisions] = React.useState(null)

    const revisionsRef = React.useRef(revisions ? revisions : [])

    const [err, setErr] = React.useState(null)

    const [eventSource, setEventSource] = React.useState(null)

    React.useEffect(() => {
        if (eventSource === null) {
            // setup event listener 
            let listener = new EventSourcePolyfill(`${url}functions/namespaces/${namespace}/tree${path}?op=function-revisions&svn=${service}&version=${version}`, {
                headers: apiKeyHeaders(apikey)
            })

            listener.onerror = (e) => {
                if (e.status === 404) {
                    setErr(e.statusText)
                } else if (e.status === 403) {
                    setErr("permission denied")
                }
            }

            async function readData(e) {
                let revs = revisionsRef.current
                if (e.data === "") {
                    return
                }
                let json = JSON.parse(e.data)
                switch (json.event) {
                    case "DELETED":
                        for (var i = 0; i < revs.length; i++) {
                            if (revs[i].name === json.revision.name) {
                                revs.splice(i, 1)
                                revisionsRef.current = revs
                                break
                            }
                        }
                        if (revs.length === 0) {
                            navigate(-1)
                        }
                        break
                    case "MODIFIED":
                        for (i = 0; i < revs.length; i++) {
                            if (revs[i].name === json.revision.name) {
                                revs[i] = json.revision
                                revisionsRef.current = revs
                                break
                            }
                        }
                        break
                    default:
                        let found = false
                        for (i = 0; i < revs.length; i++) {
                            if (revs[i].name === json.revision.name) {
                                found = true
                                break
                            }
                        }
                        if (!found) {
                            revs.push(json.revision)
                            revisionsRef.current = revs
                        }
                }

                setRevisions(JSON.parse(JSON.stringify(revisionsRef.current)))
            }

            listener.onmessage = e => readData(e)
            setEventSource(listener)
        }
    }, [revisions])

    React.useEffect(() => {
        return () => {
            CloseEventSource(eventSource)
        }
    }, [eventSource])

    return {
        revisions,
        err
    }
}


/*
    useWorkflowServices is a react hook 
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace to use for the api
      - path to use for the api of the workflow
      - apikey to provide authentication of an apikey
*/
export const useDirektivWorkflowServices = (url, stream, namespace, path, apikey, ...queryParameters) => {
    const [data, dispatchData] = React.useReducer(EventStateReducer, [])
    const [err, setErr] = React.useState(null)

    const [eventSource, setEventSource] = React.useState(null)
    const { eventSourceRef } = useEventSourceCleaner(eventSource, "useDirektivWorkflowServices");

    // Store Query parameters
    const { queryString } = useQueryString(true, queryParameters)
    const [pathString, setPathString] = React.useState(null)

    // Stream Event Source Data Dispatch Handler
    React.useEffect(() => {
        const handler = setTimeout(() => {
            if (stream && pathString !== null) {
                // setup event listener 
                let listener = new EventSourcePolyfill(`${pathString}${queryString}`, {
                    headers: apiKeyHeaders(apikey)
                })

                listener.onerror = (e) => { genericEventSourceErrorHandler(e, setErr) }

                async function readData(e) {
                    if (e.data === "") {
                        return
                    }

                    let json = JSON.parse(e.data)
                    dispatchData({
                        event: json.event,
                        data: json,
                        idKey: "serviceName",
                        idNewItemKey: "function.serviceName",
                        idData: "function",
                    })
                }

                listener.onmessage = e => readData(e)
                setEventSource(listener)
            } else {
                setEventSource(null)
            }
        }, 50);

        return () => {
            clearTimeout(handler);
        };
    }, [stream, queryString, pathString])

    // Reset states when any prop that affects path is changed
    React.useEffect(() => {
        if (stream) {
            setPathString(url && namespace && path ? `${url}functions/namespaces/${namespace}/tree${SanitizePath(path)}?op=services` : null)
        } else {
            dispatchData({ event: EVENTSTATE.CLEAR })
            setPathString(url && namespace && path ? `${url}functions/namespaces/${namespace}/tree${SanitizePath(path)}?op=services` : null)
        }
    }, [stream, path, namespace, url])

    async function getWorkflowServices(...queryParameters) {
        let resp = await fetch(`${url}functions/namespaces/${namespace}/tree/${path}?op=services${ExtractQueryString(true, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey),
            method: "GET"
        })
        if (resp.ok) {
            let json = await resp.json()
            setData(json)
            return json
        } else {
            throw new Error(await HandleError('get workflow services', resp, 'listServices'))
        }
    }

    async function deleteWorkflowService(service, version, ...queryParameters) {
        let resp = await fetch(`${url}functions/namespaces/${namespace}/tree/${path}?op=delete-service&svn=${service}&version=${version}${ExtractQueryString(true, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey),
            method: "DELETE"
        })
        if (!resp.ok) {
            throw new Error(await HandleError('deleting workflow service', resp, 'deleteWorkflowService'))
        }
    }

    return {
        data,
        err,
        deleteWorkflowService,
        getWorkflowServices
    }
}