import regex from "./regex.js"

const errorResponse = {
    code: expect.anything(),
    message: expect.anything(),
}

const unauthorizedResponse = {
    message: "Unauthorized",
    status: 401,
    type: "error",
};

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

const nodeObject = {
    name: expect.stringMatching(regex.nodeNameRegex),
    path: expect.stringMatching(regex.pathRegex),
    parent: expect.stringMatching(regex.pathRegex),
    type: expect.stringMatching(regex.nodeTypeRegex),
    expandedType: expect.stringMatching(regex.nodeExtendedTypeRegex),
    readOnly: expect.anything(),
    attributes: expect.anything(),
    oid: "",
    createdAt: expect.stringMatching(regex.timestampRegex),
    updatedAt: expect.stringMatching(regex.timestampRegex),
}

export default {errorResponse, pageInfoObject, namespaceObject, nodeObject, unauthorizedResponse}
