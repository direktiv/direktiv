import React from 'react'
import {useHistory} from 'react-router-dom'


import {CaretRightFill} from 'react-bootstrap-icons'
import DeleteNamespace from './delete-namespace';


export default function HomeSidebarNamespaces(props) {
    const {namespace, fetchNamespaces} = props

    const history = useHistory()
    let liClass = "no-border"
    if (!props.isLast) {
        liClass = ""
    }

    return (
        <li title={props.namespace} onClick={() => {
            history.push(`/p/${props.namespace}`)
        }} className={liClass} style={{paddingRight: "8px"}}>
            <div style={{display: "flex"}}>
                <div className="truncate" style={{alignItems: "center", maxWidth: "200px", display: "flex"}}>
                    <div>
                        <CaretRightFill color={"#2396d8"} style={{marginRight: "8px"}}/>
                    </div>
                    <div style={{flex: "auto"}}>
                        <span style={{color: "#2396d8"}}>{props.namespace}</span>
                    </div>
                </div>
                <div id={"action-btn-namespaces" + props.namespace}
                     style={{flex: "auto", display: "flex", alignItems: "center", flexDirection: "row-reverse"}}>
                    <DeleteNamespace namespace={props.namespace} fetchNamespaces={fetchNamespaces}/>
                </div>
            </div>
        </li>
    );
}