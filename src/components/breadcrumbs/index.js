import React from 'react';
import './style.css';
import FlexBox from '../flexbox';
import { Link } from 'react-router-dom'
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

    if (!namespace){
        return ""
    }

    return(
        <FlexBox>
            <ul>
                {breadcrumbs.map((obj)=>{
                    // ignore breadcrumbs for dividers
                    console.log(obj.key)
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
            </ul>
        </FlexBox>
    );
}

export default Breadcrumbs;