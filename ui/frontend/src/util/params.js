export function QueryParams(params) {

    params = params.replace("?", "")
    let out = {};

    let tuples = params.split("&")
    for (let i = 0; i < tuples.length; i++) {
        if (tuples[i].length !== 0) {
            let keyVal = tuples[i].split("=")
            out[keyVal[0]] = keyVal[1]
        }

    }

    return out;
}