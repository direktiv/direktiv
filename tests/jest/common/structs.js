import regex from "./regex.js"

const errorResponse = {
    code: expect.anything(),
    message: expect.anything(),
}

const pageInfoObject = {
    order: expect.anything(),
    filter: expect.anything(),
    limit: expect.anything(),
    offset: expect.anything(),
    total: expect.anything(),
}

const namespaceObject = {
    name: expect.stringMatching(regex.namespaceRegex),
    oid: "",
    createdAt: expect.stringMatching(regex.timestampRegex),
    updatedAt: expect.stringMatching(regex.timestampRegex),
}

export default { errorResponse, pageInfoObject, namespaceObject }
