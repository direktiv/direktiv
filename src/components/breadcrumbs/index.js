import React from 'react';
import './style.css';
import FlexBox from '../flexbox';
import { Link } from 'react-router-dom'
import useBreadcrumbs from 'use-react-router-breadcrumbs'

function Breadcrumbs(props) {

    const breadcrumbs = useBreadcrumbs()
    console.log(breadcrumbs)
    return(
        <FlexBox>
            <ul>
                {breadcrumbs.map((obj)=>{
                    return(
                        <li>
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