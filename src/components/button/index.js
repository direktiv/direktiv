import React from 'react';

function Button(props){

    let {children, onClick, style, className} = props;

    return(
        <div onClick={onClick} className={"btn " + className} style={style}>
            {children}
        </div>
    );
}

export default Button;