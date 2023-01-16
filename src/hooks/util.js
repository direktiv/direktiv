import _fetch from 'isomorphic-fetch'
import * as React from 'react'

// Config default config to test with 
export const Config = {
    namespace: process.env.NAMESPACE,
    url: process.env.API_URL,
    registry: "https://docker.io",
    "direktiv-token": "testapikey",
    secret: "test-secret",
    secretdata: "test-secret-data"
}

export function SanitizePath(path) {
    if (path === "") {
        return path
    }

    if (path === "/") {
        return ""
    }

    if (path.startsWith("/")) {
        return path
    }

    return "/" + path
}

// CloseEventSource closes the event source when the component unmounts
export async function CloseEventSource(eventSource) {
    if (eventSource !== null) {
        eventSource.close()
    }
}

export function TrimPathSlashes(path) {
    path.replace(/^\//, "");
    path.replace(/\/^/, "");
    return path
}

// HandleError returns a helpful message back based on the response
export async function HandleError(summary, resp, perm) {
    const contentType = resp.headers.get('content-type')

    if (resp.status === 405) {
        return `${summary}: method is not allowed`
    }

    if (resp.status !== 403) {
        if (!contentType || !contentType.includes('application/json')) {
            let text = await resp.text()
            return `${summary}: ${text}`
        } else {
            if (resp.headers.get('grpc-message')) {
                return `${summary}: ${resp.headers.get('grpc-message')}`
            } else {
                let text = (await resp.json()).message
                return `${summary}: ${text}`
            }
        }
    } else {
        // if no permission is provided, return the error message from the reponse
        if (!perm) {
          try {
            return await resp.text();
          } catch (error) {
            return `You do not have permission to '${summary}'`
          }
        }
        return `You do not have permission to '${summary}', contact system admin to grant '${perm}'`
    }
}

export function ExtractQueryString(appendMode, ...queryParameters) {
    if (queryParameters === undefined || queryParameters.length === 0) {
        return ""
    }

    let queryString = ""
    for (let i = 0; i < queryParameters.length; i++) {
        const query = queryParameters[i];
        if (i > 0 || appendMode) {
            queryString += "&" + query
        } else {
            queryString += query
        }
    }

    if (appendMode) {
        return queryString
    }

    return `?${queryString}`
}

export function QueryStringsContainsQuery(containQuery, ...queryParameters) {
    if (queryParameters === undefined || queryParameters.length === 0) {
        return false
    }

    for (let i = 0; i < queryParameters.length; i++) {
        const query = queryParameters[i];
        if (query.startsWith(`${containQuery}=`)) {
            return true
        }
    }

    return false
}

export const STATE = {
    UPDATE: 'update',
    PUSHITEM: "pushItem",
    APPENDLIST: "appendList",
    UPDATEKEY: "updateKey"
};

export const EVENTSTATE = {
    ADDED: 'ADDED',
    MODIFIED: "MODIFIED",
    DELETED: "DELETED",

    CLEAR: "CLEAR"
};

export function EventStateReducer(state, action) {

    // Clear state
    if (action.event === EVENTSTATE.CLEAR){
        return []
    }

    // Get unique id of new item
    const newItemID = getPropStr(action.data, action.idNewItemKey)
    if (!newItemID) {
        return state
    }

    // Check if unique id already exists and track its itemIndex if it does.
    let itemIndex = -1
    const newState = JSON.parse(JSON.stringify(state))
    for (let i = 0; i < newState.length; i++) {
        const stateItemID = getPropStr(newState[i], action.idKey)
        if (stateItemID === newItemID) {
            itemIndex = i
            break;
        }
    }

    switch (action.event) {
        case EVENTSTATE.MODIFIED:
            if (itemIndex >= 0) {
                if (action.idData) {
                    const newItem = getPropStr(action.data, action.idData)
                    newState[itemIndex] = {...newItem}
                } else {
                    newState[itemIndex] = {...action.data}
                }
            }
            break
        case EVENTSTATE.DELETED:
            if (itemIndex >= 0) {
                newState.splice(itemIndex, 1)
            }

            break
        case EVENTSTATE.ADDED:
        default:
            if (itemIndex === -1) {
                if (action.idData) {
                    const newItem = getPropStr(action.data, action.idData)
                    newState.push(newItem)
                } else {
                    newState.push(action.data)
                }
            }

    }

    return newState
}

const getPropStr = (object, pathStr) => {
    const path = pathStr.split(".")

    try {
        return getProp(object, path)
    } catch (err) {
        return null
    }
}

const getProp = (object, path) => {
    if (path.length === 1) return object[path[0]];
    else if (path.length === 0) throw error;
    else {
        if (object[path[0]]) return getProp(object[path[0]], path.slice(1));
        else {
            object[path[0]] = {};
            return getProp(object[path[0]], path.slice(1));
        }
    }
};

export function StateReducer(state, action) {
    switch (action.type) {
        case STATE.UPDATE:
            return action.data;
        case STATE.PUSHITEM:
            let pushListData = state ? JSON.parse(JSON.stringify(state)) : []
            pushListData.push(action.data)
            return pushListData
        case STATE.APPENDLIST:
            let appendListData = state ? JSON.parse(JSON.stringify(state)) : []
            for (let i = 0; i < action.data.length; i++) {
                appendListData.push(action.data[i])
            }

            return appendListData
        case STATE.UPDATEKEY:
            if (state[action.key]) {
                state[action.key] = JSON.parse(JSON.stringify(action.data))
            }

            return state
        default:
            return state
    }
}

// Auto clean eventsource when changed or unmounted
export const useEventSourceCleaner = (eventSource, extra) => {
    const eventSourceRef = React.useRef(eventSource);

    // CLEANUP: close old eventsource and updates ref
    React.useEffect(() => {
        eventSourceRef.current = eventSource

        return () => {
            CloseEventSource(eventSource)
        }
    }, [eventSource])

    // CLEANUP: close eventsource on umount
    React.useEffect(() => {
        return () => {
            CloseEventSource(eventSourceRef.current)
        }
    }, [])

    return {
        eventSourceRef
    }
}

// Handle changes to queryParameters and return new query string when changed
// throttle can be used to control how frequently to update queryString in ms. Default = 50
export const useQueryString = (appendMode, queryParameters, throttle) => {
    const [queryString, setQueryString] = React.useState("")

    React.useEffect(() => {
        let newQueryString = ExtractQueryString(appendMode, ...queryParameters)
        if (newQueryString !== queryString) {
            setQueryString(newQueryString)
        }
    }, [appendMode, queryParameters, queryString, throttle])

    return {
        queryString
    }
}

export const genericEventSourceErrorHandler = (error, setError) => {
    if (error.status === 404) {
        setError(error.statusText)
    } else if (error.status === 403) {
        setError("permission denied")
    }
}

//  isValueValid : Checks if value is not undefined and not null
export const isValueValid = (value) => {
    if (typeof value !== "undefined" && value !== null) {
        return true
    } else {
        return false
    }
}

//  apiKeyHeaders : Returns header object with "direktiv-token" set to apiKey if key has a valid value.
//  An empty object is returned otherwise
export const apiKeyHeaders = (apiKey) => {
    if (isValueValid(apiKey)) {
        const isBearer = apiKey.length > 200;
        if (isBearer) {
          return {
            Authorization: `Bearer ${apiKey}`,
          };
        }
        return {
          "direktiv-token": apiKey,
        };
    }
    return {}
}