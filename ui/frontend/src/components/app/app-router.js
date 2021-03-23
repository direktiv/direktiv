import React, {useCallback, useContext, useState} from "react";

import "css/app.css";

import {BrowserRouter as Router, Route, Switch} from "react-router-dom";

import "bootstrap/dist/css/bootstrap.min.css";
import "animate.css";

import Container from "react-bootstrap/Container";

import TopNav from "components/app/header";
import Home from "components/home/home";
import ResourceNotFound from "util/not-found";
import Workflow from "components/workflow/workflow";
import Breadcrumbs from "components/app/breadcrumbs";
import Namespace from "components/namespaces/namespace";
import CustomToast from "components/app/toasts";
import Instance from "components/instances/instance";
import Footer from "components/app/footer";

import ServerContext from "components/app/context";

export default function AppRouter(props) {
    const context = useContext(ServerContext);
    const [toasts, setToasts] = useState([]);
    const [showToasts, setShowToasts] = useState(true);

    function addToast(toastID, body) {
        if (showToasts) {
            setToasts([
                ...toasts,
                {title: toastID, body: body, finished: false, timeStamp: Date.now()},
            ]);
        }
    }

    const Fetch = useCallback(
        (path, opts) => {
            return fetch(`${context.SERVER_BIND}${path}`, opts);
        },
        []
    );

    function toastCount() {
        if (toasts) {
            return toasts.length;
        }
        return 0;
    }

    function clearToasts() {
        setToasts((toasts) => {
            return [];
        });
    }

    function toggleToasts() {
        setShowToasts((showToasts) => {
            return !showToasts;
        });
    }

    const handleRemoveItem = React.useCallback((timeStamp) => {
        setToasts((toasts) => {
            let cleanup = true;
            for (var i = 0; i < toasts.length; i++) {
                if (toasts[i].timeStamp === timeStamp) {
                    toasts[i].finished = true;
                }

                if (!toasts[i].finished) {
                    cleanup = false;
                }
            }

            if (cleanup) {
                return [];
            }
            return toasts;
        });
    }, []);

    let toastElements = [];
    for (let i = 0; i < toasts.length; i++) {
        if (!toasts[i].finished) {
            toastElements.push(
                <CustomToast
                    key={`toast-${i}`}
                    title={toasts[i].title}
                    destroyCallback={handleRemoveItem}
                    timeStamp={toasts[i].timeStamp}
                    body={toasts[i].body}
                />
            );
        }
    }

    return (
        <>
            <ServerContext.Provider
                value={{
                    ...context,
                    Fetch: Fetch,
                    AddToast: addToast,
                    showToasts: showToasts,
                    ToggleToast: toggleToasts,
                    ToastCount: toastCount,
                    ClearToasts: clearToasts,
                }}
            >
                <Router>
                    <Container fluid id="primary-container">
                        <TopNav
                        />
                        <Container style={{padding: "16px 0px 8px 0px", flex: "auto", minWidth: "350px"}}>
                            <div
                                style={{
                                    position: "fixed",
                                    bottom: "0%",
                                    right: "0%",
                                    padding: "0.5em 0.5em 0.5em 0.5em ",
                                    zIndex: "10000",
                                }}
                            >
                                {toastElements}
                            </div>
                            <Breadcrumbs/>
                            <Switch>
                                <Route exact path="/p/:namespace/w/:workflow" component={Workflow}/>
                                <Route exact path="/p/:namespace" component={Namespace}/>
                                <Route exact path="/i/:namespace/:workflow/:id" component={Instance}/>
                                <Route exact path="/" component={Home}/>
                                <Route exact path="/404" component={ResourceNotFound}/>
                            </Switch>
                            <div style={{color: "#f5f5f5"}}>{context.REACT_APP_GIT_HASH}</div>
                        </Container>
                        <Footer/>
                    </Container>
                </Router>
            </ServerContext.Provider>
        </>
    );
}
