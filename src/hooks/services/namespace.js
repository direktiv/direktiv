import * as React from 'react'
import { CloseEventSource, HandleError, ExtractQueryString, apiKeyHeaders } from '../util'
const fetch = require("isomorphic-fetch")
const { EventSourcePolyfill } = require('event-source-polyfill')

/* 
    useNamespaceServiceRevision takes
    - url
    - namespace
    - service
    - revision
    - apikey
*/
export const useDirektivNamespaceServiceRevision = (url, namespace, service, revision, apikey) => {
    const [revisionDetails, setRevisionDetails] = React.useState(null)
    const [podSource, setPodSource] = React.useState(null)
    const [pods, setPods] = React.useState([])
    const [err, setErr] = React.useState(null)
    const [revisionSource, setRevisionSource] = React.useState(null)

    const podsRef = React.useRef(pods)


    React.useEffect(() => {
        if (podSource === null) {
            let listener = new EventSourcePolyfill(`${url}functions/namespaces/${namespace}/function/${service}/revisions/${revision}/pods`, {
                headers: apiKeyHeaders(apikey)
            })

            listener.onerror = (e) => {
                if (e.status === 404) {
                    setErr(e.statusText)
                } else if (e.status === 403) {
                    setErr("permission denied")
                } else {
                    try {
                        let json = JSON.parse(e.data)
                        setErr(json.Message)
                    } catch (e) {
                        // TODO
                    }
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
    }, [apikey])

    React.useEffect(() => {
        if (revisionSource === null) {
            // setup event listener 
            let listener = new EventSourcePolyfill(`${url}functions/namespaces/${namespace}/function/${service}/revisions/${revision}`, {
                headers: apiKeyHeaders(apikey)
            })

            listener.onerror = (e) => {
                if (e.status === 404) {
                    setErr(e.statusText)
                } else if (e.status === 403) {
                    setErr("permission denied")
                } else {
                    try {
                        let json = JSON.parse(e.data)
                        setErr(json.Message)
                    } catch (e) {
                        // TODO handle error/error handling
                    }
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
    }, [revisionSource, apikey])

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
    useNamespaceService takes
    - url
    - namespace
    - service
    - navigate (react router object to navigate backwards)
    - apikey
*/
export const useDirektivNamespaceService = (url, namespace, service, navigate, apikey) => {
    const [revisions, setRevisions] = React.useState(null)
    const [fn, setFn] = React.useState(null)
    const [traffic, setTraffic] = React.useState(null)
    const [config, setConfig] = React.useState(null)
    const revisionsRef = React.useRef(revisions ? revisions : [])


    const [err, setErr] = React.useState(null)

    const [trafficSource, setTrafficSource] = React.useState(null)
    const [eventSource, setEventSource] = React.useState(null)

    React.useEffect(() => {
        if (trafficSource === null) {
            // setup event listener 
            let listener = new EventSourcePolyfill(`${url}functions/namespaces/${namespace}/function/${service}`, {
                headers: apiKeyHeaders(apikey)
            })

            listener.onerror = (e) => {
                if (e.status === 404) {
                    setErr(e.statusText)
                } else if (e.status === 403) {
                    setErr("permission denied")
                } else {
                    try {
                        let json = JSON.parse(e.data)
                        setErr(json.Message)
                    } catch (e) {
                        // TODO
                    }
                }
            }

            async function readData(e) {
                if (e.data === "") {
                    return
                }
                let json = JSON.parse(e.data)

                if (json.event === "MODIFIED" || json.event === "ADDED") {
                    setFn(JSON.parse(JSON.stringify(json.function)))
                    setTraffic(JSON.parse(JSON.stringify(json.traffic)))
                }
            }

            listener.onmessage = e => readData(e)
            setTrafficSource(listener)
        }
    }, [fn, apikey])

    React.useEffect(() => {
        if (eventSource === null) {
            // setup event listener 
            let listener = new EventSourcePolyfill(`${url}functions/namespaces/${namespace}/function/${service}/revisions`, {
                headers: apiKeyHeaders(apikey)
            })

            listener.onerror = (e) => {
                if (e.status === 404) {
                    setErr(e.statusText)
                } else if (e.status === 403) {
                    setErr("permission denied")
                } else {
                    try {
                        let json = JSON.parse(e.data)
                        setErr(json.Message)
                    } catch (e) {
                        // TODO
                    }
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
                revisionsRef.current.sort(function (a, b) {
                    return parseInt(a.generation) < parseInt(b.generation) ? 1 : -1;
                })
                setRevisions(JSON.parse(JSON.stringify(revisionsRef.current)))
            }

            listener.onmessage = e => readData(e)
            setEventSource(listener)
        }
    }, [revisions, apikey])

    React.useEffect(() => {
        return () => {
            CloseEventSource(eventSource)
            CloseEventSource(trafficSource)
        }
    }, [eventSource, trafficSource])

    async function getNamespaceServiceConfig(...queryParameters) {
        let resp = await fetch(`${url}functions/namespaces/${namespace}/function/${service}${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey),
            method: "GET"
        })
        if (resp.ok) {
            let json = await resp.json()
            setConfig(json.config)
            return json.config
        } else {
            throw new Error(await HandleError('get namespace service', resp, 'getService'))
        }

    }

    async function createNamespaceServiceRevision(image, minScale, size, cmd, traffic, ...queryParameters) {
        let resp = await fetch(`${url}functions/namespaces/${namespace}/function/${service}${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey),
            method: "POST",
            body: JSON.stringify({
                trafficPercent: traffic,
                cmd,
                image,
                minScale,
                size
            })
        })
        if (!resp.ok) {
            throw new Error(await HandleError('create namespace service revision', resp))
        }
    }

    async function deleteNamespaceServiceRevision(rev, ...queryParameters) {
        let resp = await fetch(`${url}functions/namespaces/${namespace}/function/${service}/revisions/${rev}${ExtractQueryString(false, ...queryParameters)}`, {
            method: "DELETE",
            headers: apiKeyHeaders(apikey),
        })
        if (!resp.ok) {
            throw new Error(await HandleError('delete namespace service revision', resp, 'deleteRevision'))
        }
    }

    async function setNamespaceServiceRevisionTraffic(rev1, rev1value, rev2, rev2value, ...queryParameters) {
        let trafficarr = []
        if (rev1 !== "") {
            trafficarr.push({
                revision: rev1,
                percent: rev1value
            })
        }
        if (rev2 !== "") {
            trafficarr.push({
                revision: rev2,
                percent: rev2value
            })
        }
        let resp = await fetch(`${url}functions/namespaces/${namespace}/function/${service}${ExtractQueryString(false, ...queryParameters)}`, {
            method: "PATCH",
            body: JSON.stringify({ values: trafficarr }),
            headers: apiKeyHeaders(apikey)
        })
        if (!resp.ok) {
            throw new Error(await HandleError('update traffic namespace service', resp, 'updateTraffic'))
        }
    }

    return {
        revisions,
        fn,
        config,
        traffic,
        err,
        createNamespaceServiceRevision,
        deleteNamespaceServiceRevision,
        getNamespaceServiceConfig,
        setNamespaceServiceRevisionTraffic
    }
}
/*
    useNamespaceServices is a react hook 
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace to use for the api
      - apikey to provide authentication of an apikey
*/
export const useDirektivNamespaceServices = (url, stream, namespace, apikey) => {
    const [data, setData] = React.useState(null)
    const [config, setConfig] = React.useState(null)

    const functionsRef = React.useRef(data ? data : [])
    const [err, setErr] = React.useState(null)
    const [eventSource, setEventSource] = React.useState(null)


    React.useEffect(() => {
        if (stream) {
            if (eventSource === null) {
                // setup event listener 
                let listener = new EventSourcePolyfill(`${url}functions/namespaces/${namespace}`, {
                    headers: apiKeyHeaders(apikey)
                })

                listener.onerror = (e) => {
                    if (e.status === 404) {
                        setErr(e.statusText)
                    } else if (e.status === 403) {
                        setErr("permission denied")
                    } else {
                        try {
                            let json = JSON.parse(e.data)
                            setErr(json.Message)
                        } catch (e) {
                            // TODO
                        }
                    }
                }

                async function readData(e) {
                    let funcs = functionsRef.current
                    if (e.data === "") {
                        return
                    }
                    let json = JSON.parse(e.data)
                    switch (json.event) {
                        case "DELETED":
                            for (var i = 0; i < funcs.length; i++) {
                                if (funcs[i].serviceName === json.function.serviceName) {
                                    funcs.splice(i, 1)
                                    functionsRef.current = funcs
                                    break
                                }
                            }
                            break
                        case "MODIFIED":
                            for (i = 0; i < funcs.length; i++) {
                                if (funcs[i].serviceName === json.function.serviceName) {
                                    funcs[i] = json.function
                                    functionsRef.current = funcs
                                    break
                                }
                            }
                            break
                        default:
                            let found = false
                            for (i = 0; i < funcs.length; i++) {
                                if (funcs[i].serviceName === json.function.serviceName) {
                                    found = true
                                    break
                                }
                            }
                            if (!found) {
                                funcs.push(json.function)
                                functionsRef.current = funcs
                            }
                    }
                    setData(JSON.parse(JSON.stringify(functionsRef.current)))
                }

                listener.onmessage = e => readData(e)
                setEventSource(listener)
            }
        } else {
            if (data === null) {
                getNamespaceServices()
            }
        }
    }, [data, apikey])

    React.useEffect(() => {
        return () => CloseEventSource(eventSource)
    }, [eventSource])


    async function getNamespaceServices(...queryParameters) {
        let resp = await fetch(`${url}functions/namespaces/${namespace}${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey),
            method: "GET"
        })
        if (resp.ok) {
            let json = await resp.json()
            setData(json.functions)
            return json.functions
        } else {
            throw new Error(await HandleError('get namespace service', resp, 'listServices'))
        }
    }

    async function getNamespaceConfig(...queryParameters) {
        let resp = await fetch(`${url}functions/namespaces/${namespace}${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey),
            method: "GET"
        })
        if (resp.ok) {
            let json = await resp.json()
            setConfig(json.config)
            return json.config
        } else {
            throw new Error(await HandleError('get namespace service', resp, 'listServices'))
        }
    }

    async function createNamespaceService(name, image, minScale, size, cmd, ...queryParameters) {
        let resp = await fetch(`${url}functions/namespaces/${namespace}${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey),
            method: "POST",
            body: JSON.stringify({
                cmd,
                image,
                minScale,
                name,
                size
            })
        })
        if (!resp.ok) {
            throw new Error(await HandleError('create namespace service', resp))
        }
    }

    async function deleteNamespaceService(name, ...queryParameters) {
        let resp = await fetch(`${url}functions/namespaces/${namespace}/function/${name}${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey),
            method: "DELETE"
        })
        if (!resp.ok) {
            throw new Error(await HandleError('delete namespace service', resp, 'deleteService'))
        }
    }


    return {
        data,
        err,
        config,
        getNamespaceConfig,
        getNamespaceServices,
        createNamespaceService,
        deleteNamespaceService
    }

}