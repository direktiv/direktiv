import React from 'react';
import './style.css';

function Button(props){

    let {children, onClick, style, className, title} = props;
    let loader = (<></>);

    if (className) {
        if (className.includes("btn-loading")) {
            loader = (
                <div className="btn-loader" />
            )
        }
    }

    return(
        <div onClick={onClick} className={"btn " + className} style={style} title={title}>
            {children}
            {loader}
        </div>
    );
}

export default Button;