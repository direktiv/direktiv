import * as React from 'react'
import { CloseEventSource, HandleError, ExtractQueryString, apiKeyHeaders } from '../util'
const { EventSourcePolyfill } = require('event-source-polyfill')
const fetch = require('isomorphic-fetch')

/*
    useNamespaces is a react hook which returns createNamespace, deleteNamespace and data
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - apikey to provide authentication of an apikey
*/
export const useDirektivNamespaces = (url, stream, apikey, ...queryParameters) => {

    const [data, setData] = React.useState(null)
    const [load, setLoad] = React.useState(true)
    const [err, setErr] = React.useState(null)
    const [eventSource, setEventSource] = React.useState(null)

    // Store Query parameters
    const [queryString, setQueryString] = React.useState(ExtractQueryString(false, ...queryParameters))

    // Stores PageInfo about namespace list stream
    const [pageInfo, setPageInfo] = React.useState(null)

    React.useEffect(() => {
        if (stream) {
            if (eventSource === null) {
                // setup event listener 
                let listener = new EventSourcePolyfill(`${url}namespaces${queryString}`, {
                    headers: apiKeyHeaders(apikey)
                })

                listener.onerror = (e) => {
                    setErr(e)
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
                    setData(json.results)
                    setPageInfo(json.pageInfo)
                }

                listener.onmessage = e => readData(e)
                setEventSource(listener)
                setLoad(false)
                setErr("")
            }
        } else {
            if (data === null) {
                getNamespaces()
            }
        }
    }, [data, apikey])

    React.useEffect(() => {
        if (!load && eventSource !== null) {
            CloseEventSource(eventSource)
            // setup event listener 
            let listener = new EventSourcePolyfill(`${url}namespaces${queryString}`, {
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
                setData(json.results)
                setPageInfo(json.pageInfo)
            }

            listener.onmessage = e => readData(e)
            setEventSource(listener)
            setErr("")
        }
    }, [apikey])

    // If queryParameters change and streaming: update queryString, and reset sse connection
    React.useEffect(() => {
        if (stream) {
            let newQueryString = ExtractQueryString(false, ...queryParameters)
            if (newQueryString !== queryString) {
                setQueryString(newQueryString)
                CloseEventSource(eventSource)
                setEventSource(null)
            }
        }
    }, [eventSource, queryParameters, queryString, stream])

    React.useEffect(() => {
        return () => {
            CloseEventSource(eventSource)
        }
    }, [eventSource])

    // getNamespaces returns a list of namespaces
    async function getNamespaces(...queryParameters) {
        // fetch namespace list by default
        let resp = await fetch(`${url}namespaces${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (resp.ok) {
            let json = await resp.json()
            setData(json.results)
            setPageInfo(json.pageInfo)
            return json.results
        } else {
            throw new Error((await HandleError('list namespaces', resp, 'listNamespaces')))
        }
    }

    // createNamespace creates a namespace from direktiv
    async function createNamespace(namespace, ...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}${ExtractQueryString(false, ...queryParameters)}`, {
            method: "PUT",
            headers: apiKeyHeaders(apikey)
        })
        if (!resp.ok) {
            throw new Error(await HandleError('create a namespace', resp, 'addNamespace'))
        }
    }

    async function createMirrorNamespace(namespace, mirrorSettings, ...queryParameters) {
        let request = {
            method: "PUT",
            body: JSON.stringify(mirrorSettings),
            headers: apiKeyHeaders(apikey)
        }

        let resp = await fetch(`${url}namespaces/${namespace}${ExtractQueryString(false, ...queryParameters)}`, request)
        if (!resp.ok) {
            throw new Error(await HandleError('create a mirror namespace', resp, 'addNamespace'))
        }
    }

    // deleteNamespace deletes a namespace from direktiv
    async function deleteNamespace(namespace, ...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}?recursive=true${ExtractQueryString(true, ...queryParameters)}`, {
            method: "DELETE",
            headers: apiKeyHeaders(apikey)
        })
        if (!resp.ok) {
            throw new Error(await HandleError('delete a namespace', resp, 'deleteNamespace'))
        }
    }

    return {
        data,
        err,
        pageInfo,
        createNamespace,
        deleteNamespace,
        getNamespaces,
        createMirrorNamespace
    }
}
