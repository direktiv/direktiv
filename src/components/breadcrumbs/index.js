import React from 'react';
import './style.css';
import FlexBox from '../flexbox';

function Breadcrumbs(props) {
    return(
        <FlexBox>
            <ul>
                <li>
                    <a href="">
                        Direktiv
                    </a>
                </li>
                <li>
                    <a href="">
                        Example
                    </a>
                </li>
            </ul>
        </FlexBox>
    );
}

export default Breadcrumbs;