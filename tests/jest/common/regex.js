const timestampRegex = `^2.*Z$`
const namespaceRegex = String.raw`^(([a-z][a-z0-9_\-\.]*[a-z0-9])|([a-z]))$`
const nodeNameRegex = String.raw`(^$)|(^(([a-z][a-z0-9_\-\.]*[a-z0-9])|([a-z]))$)`
const pathRegex = String.raw`(^[\/]((([a-z][a-z0-9_\-\.]*[a-z0-9])|([a-z]))[\/]?)*$)`
const nodeTypeRegex = String.raw`^directory$`
const nodeExtendedTypeRegex = String.raw`^directory$`


export default {
    timestampRegex,
    namespaceRegex,
    nodeNameRegex,
    pathRegex,
    nodeTypeRegex,
    nodeExtendedTypeRegex,
}