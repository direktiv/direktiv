const timestampRegex = `^2.*Z$`
const namespaceRegex = String.raw`^(([a-z][a-z0-9_\-\.]*[a-z0-9])|([a-z]))$`

export default {
    timestampRegex,
    namespaceRegex
}