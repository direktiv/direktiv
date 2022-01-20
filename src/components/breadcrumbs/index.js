import React from 'react';
import './style.css';
import FlexBox from '../flexbox';
import { Link, useSearchParams } from 'react-router-dom'
import useBreadcrumbs from 'use-react-router-breadcrumbs'
import {GenerateRandomKey} from '../../util';

const routes = [
    { path: '/jq', breadcrumb: 'JQ Playground' },
    { path: '/g/services', breadcrumb: 'Global Services' },
    { path: '/g/registries', breadcrumb: 'Global Registries'},
    { path: '/n/:namespace/services', breadcrumb: "Namespace Services"}
];

function Breadcrumbs(props) {
    const {namespace} = props
    const breadcrumbs = useBreadcrumbs(routes)
    const [searchParams] = useSearchParams() // removed 'setSearchParams' from square brackets (this should not affect anything: search 'destructuring assignment')
    
    if (!namespace){
        return <></>
    }

    return(
        <FlexBox>
            <ul>
                {breadcrumbs.length < 9 ? breadcrumbs.map((obj)=>{
                    // ignore breadcrumbs for dividers
                    if(obj.key === "/g" || obj.key === "/n" || obj.key === "/" || obj.key === `/n/${namespace}/explorer` ) {
                        return ""
                    }
                    let key = GenerateRandomKey("crumb-");
               
                    return(
                        <li id={key} key={key}>
                            <Link to={obj.key}>
                                {obj.breadcrumb}
                            </Link>
                        </li>
                    )
                }) : 
               <>
               {breadcrumbs.slice(0, 6).map((obj)=>{
                    if(obj.key === "/g" || obj.key === "/n" || obj.key === "/" || obj.key === `/n/${namespace}/explorer` ) {
                        return ""
                    }
                    let key = GenerateRandomKey("crumb-");
               
                    return(
                        <li id={key} key={key}>
                            <Link to={obj.key}>
                                {obj.breadcrumb}
                            </Link>
                        </li>
                    )
                })}
                <li id={"crumb-divider"} key={"crumb-divider"}>
                    <span>{". . . "}</span>
                </li>
                {breadcrumbs.slice(-3).map((obj)=>{
                    let key = GenerateRandomKey("crumb-");
               
                    return(
                        <li id={key} key={key}>
                            <Link to={obj.key}>
                                {obj.breadcrumb}
                            </Link>
                        </li>
                    )
                })
                }
               </>
                }
                
                
                {searchParams.get("function") && searchParams.get("version") ? 
                    <li id={`${searchParams.get("function")}-${searchParams.get("version")}`} key={`${searchParams.get("function")}-${searchParams.get("version")}`}>
                        <Link to={`${window.location.pathname}?function=${searchParams.get("function")}&version=${searchParams.get("version")}`}>
                            {searchParams.get("function")}
                        </Link>
                    </li>
                    :""
                }
                { searchParams.get("revision") && searchParams.get("function") && searchParams.get("version") ? 
                    <li id={`${searchParams.get("function")}-${searchParams.get("version")}-${searchParams.get("revision")}`} key={`${searchParams.get("function")}-${searchParams.get("version")}-${searchParams.get("revision")}`}>
                       <Link to={`${window.location.pathname}?function=${searchParams.get("function")}&version=${searchParams.get("version")}&revision=${searchParams.get("revision")}`}>
                           {searchParams.get("revision")}
                       </Link>
                   </li>
                :""}
            </ul>
        </FlexBox>
    );
}

export default Breadcrumbs;