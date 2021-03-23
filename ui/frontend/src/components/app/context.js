import React from "react";

let ServerContext = React.createContext({
    // dev server bind
    SERVER_BIND: "http://localhost:8080/api",
});

if (process.env.NODE_ENV === "production") {
    ServerContext = React.createContext({
        SERVER_BIND: "/api"
    });
}

export default ServerContext;
