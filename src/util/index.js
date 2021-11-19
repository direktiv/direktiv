export const Config = {
    url: process.env.REACT_APP_API
}

export function GenerateRandomKey(prefix) {
    if (!prefix) {
        prefix = "";
    }

    return prefix + Array(16).fill().map(()=>"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789".charAt(Math.random()*62)).join("")
}

