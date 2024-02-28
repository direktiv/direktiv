const timestampRegex = String.raw`^((?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?)$`
const namespaceRegex = String.raw`^(([a-z][a-z0-9_\-\.]*[a-z0-9])|([a-z]))$`
const nodeNameRegex = String.raw`(^$)|(^(([a-z][a-z0-9_\-\.]*[a-z0-9])|([a-z]))$)`
const pathRegex = String.raw`(^[\/]((([a-z][a-z0-9_\-\.]*[a-z0-9])|([a-z]))[\/]?)*$)`
const nodeTypeRegex = String.raw`^directory$`
const nodeExtendedTypeRegex = String.raw`^directory$`

const hashRegex = String.raw`^[0-9a-f]{64}$`
const base64Regex = String.raw`^[-A-Za-z0-9+/=]*$`
const uuidRegex = String.raw`^[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}$`

export default {
	timestampRegex,
	namespaceRegex,
	nodeNameRegex,
	pathRegex,
	nodeTypeRegex,
	nodeExtendedTypeRegex,

	hashRegex,
	base64Regex,
	uuidRegex,
}
