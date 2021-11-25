import React from 'react';
import './style.css';

function Button(props){

    let {children, onClick, style, className} = props;
    let loader = (<></>);

    if (className) {
        if (className.includes("btn-loading")) {
            loader = (
                <div className="btn-loader" />
            )
        }
    }

    return(
        <div onClick={onClick} className={"btn " + className} style={style}>
            {children}
            {loader}
        </div>
    );
}

export default Button;