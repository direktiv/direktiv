import React from 'react'
import 'css/home-sidebar.css'


import {Journals} from 'react-bootstrap-icons'


import HomeSidebarNamespaces from 'components/home/home-sidebar-namespaces'
import NewNamespace from './new-namespace'


export default function HomepageSidebar(props) {
    const {namespaces, fetchNamespaces} = props

    let p = [];
    for (let i = 0; i < namespaces.length; i++) {
        let isLast = false
        if (i === namespaces.length - 1) {
            isLast = true
        }
        p.push(
            <HomeSidebarNamespaces fetchNamespaces={fetchNamespaces} key={i} isLast={isLast}
                                   namespace={namespaces[i].name}/>
        )
    }

    return (
        <div id="home-sidebar">
            <ul>
                <li className="title">
                    <div className="truncate" style={{display: "flex", alignItems: "center", paddingRight: "3px"}}>
                        <div style={{flex: "auto", fontSize: "14px"}}>
                            <Journals style={{marginRight: "8px"}}/>
                            Namespaces
                        </div>
                        <NewNamespace fetchNamespaces={fetchNamespaces}/>
                    </div>
                </li>
                {p.length > 0 ?
                    p
                    :
                    <div style={{padding: "5px", fontStyle: "italic", color: "#b5b5b5"}}>
                        Click the + to create a namespace
                    </div>
                }
            </ul>
        </div>
    );
}
