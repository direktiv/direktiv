import regex from "./regex.js"

const namespaceObject = {
    name: expect.stringMatching(regex.namespaceRegex),
    oid: "",
    createdAt: expect.stringMatching(regex.timestampRegex),
    updatedAt: expect.stringMatching(regex.timestampRegex),
}

export default { namespaceObject }
