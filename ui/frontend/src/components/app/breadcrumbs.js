import React, {useContext} from 'react'
import {Link, matchPath, useLocation} from 'react-router-dom'
import 'css/breadcrumbs.css'

import {Archive, ArchiveFill, Bell, BellFill, House} from 'react-bootstrap-icons'

import Breadcrumb from 'react-bootstrap/Breadcrumb'
import Tooltip from 'react-bootstrap/Tooltip'
import OverlayTrigger from 'react-bootstrap/OverlayTrigger'

import ServerContext from 'components/app/context'


export const Breadcrumbs = () => {
    const context = useContext(ServerContext);
    let location = useLocation();

    const renderTooltipToggleToast = (props) => (
        <Tooltip id="button-tooltip" {...props}>
            Toggle New Notifications
        </Tooltip>
    );

    const renderTooltipClearToast = (props) => (
        <Tooltip id="button-tooltip" {...props}>
            Clear Notifications </Tooltip>
    );


    const renderTooltipEmptyToast = (props) => (
        <Tooltip id="button-tooltip" {...props}>
            No Notifications
        </Tooltip>
    );

    let match = matchPath(location.pathname, {
        path: "/p/:namespace/w/:workflow"
    })
    if (!match) {
        match = matchPath(location.pathname, {
            path: "/p/:namespace"
        })
    }

    let items = [];
    items.push(
        <li key="home" className="breadcrumb-item">
            <Link to="/">
                <House className="bc-svg" style={{marginRight: "8px"}}/>
                Home
            </Link>
        </li>
    )

    if (match) {
        let namespace = match.params.namespace;
        let workflow = match.params.workflow;

        if (namespace) {
            items.push(
                <li key="bc-namespace" className="breadcrumb-item">
                    <Link to={`/p/${namespace}`}>
                        {namespace}
                    </Link>
                </li>
            )
        }

        if (workflow) {
            items.push(
                <li key="bc-workflow" className="breadcrumb-item">
                    <Link to={`/p/${namespace}/w/${workflow}`}>
                        {workflow}
                    </Link>
                </li>
            )
        }
    }

    let t = context.ToastCount()

    items.push(
        <div key="notifications-crumb-toast-clear" style={{marginLeft: "auto", cursor: "pointer"}} onClick={() => {

            if (t > 0) {
                context.ClearToasts()
            }
        }}>
            <OverlayTrigger
                placement="left"
                delay={{show: 100, hide: 150}}
                overlay={t > 0 ? (renderTooltipClearToast) : (renderTooltipEmptyToast)}
            >
                <div>
                    {t > 0 ? (<ArchiveFill/>) : (<Archive/>)}
                </div>
            </OverlayTrigger>
        </div>
    )

    items.push(
        <div key="notifications-crumb-toast-toggle" style={{marginLeft: "0.55rem", cursor: "pointer"}} onClick={() => {
            context.ToggleToast()
        }}>
            <OverlayTrigger
                placement="left"
                delay={{show: 100, hide: 150}}
                overlay={renderTooltipToggleToast}
            >
                <div>
                    {context.showToasts ? (<BellFill/>) : (<Bell/>)}
                </div>
            </OverlayTrigger>
        </div>
    )

    return (
        <>
            <Breadcrumb id="breadcrumbs">
                {items}
            </Breadcrumb>
        </>

    );
}

export default Breadcrumbs;