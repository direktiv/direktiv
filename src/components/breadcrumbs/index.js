import React from 'react';
import './style.css';
import FlexBox from '../flexbox';
import { Link } from 'react-router-dom'
import useBreadcrumbs from 'use-react-router-breadcrumbs'
import {GenerateRandomKey} from '../../util';

function Breadcrumbs(props) {

    const breadcrumbs = useBreadcrumbs()

    return(
        <FlexBox>
            <ul>
                {breadcrumbs.map((obj)=>{
                    // ignore breadcrumbs for dividers
                    if(obj.key === "/g" || obj.key === "/n") {
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