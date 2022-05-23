import React from 'react';
import './style.css';

function FlexBox(props){

    let {id, children, onClick, style, className, ref, key} = props;
    if (!className) {
        className = ""
    }

    return(
        <div ref={ref} key={key} id={id} onClick={onClick} style={style} className={"flex-box " + className}>
            {children}
        </div>
    );
}

export default FlexBox;