import React from 'react';
import './style.css';
import FlexBox from '../flexbox';
import {BiChevronLeft, BiChevronRight} from 'react-icons/bi';

function Pagination(props) {

    let {max, currentIndex} = props;
    let min = currentIndex-1;
    if (min < 1) {
        min = 1
    }

    if (max === currentIndex) {
        min = max-4;
        if (min < 1) {
            min = 1;
        }
    }

    let rangeMin = min;
    if (max - rangeMin < 5) {
        rangeMin = max - 5;
    } 

    let pageBtns = [];
    for (let i = rangeMin; i < min+5; i++) {

        if (i > max) {
            break
        }

        if ((i === min+4) && (i !== max)) {
            pageBtns.push(
                <PaginationButton label="..."/>
            )
            pageBtns.push(
                <PaginationButton currentIndex={i === currentIndex} label={max} onClick={() => {
                    console.log("navigate to page " + max);
                }}/>
            )

            break;
        } else {
            pageBtns.push(
                <PaginationButton currentIndex={i === currentIndex} label={i} onClick={() => {
                    console.log("navigate to page " + i);
                }}/>
            )
        }
    }

    return(
        <FlexBox className="pagination-container auto-margin">
            <FlexBox className="pagination-btn" style={{ maxWidth: "24px" }}>
                <BiChevronLeft className="auto-margin" />
            </FlexBox>
            {pageBtns}
            <FlexBox className="pagination-btn" >
                <BiChevronRight className="auto-margin" />
            </FlexBox>
        </FlexBox>
    )
}

export default Pagination;

function PaginationButton(props) {

    let {label, onClick, currentIndex} = props;
    let classes = "pagination-btn auto-margin";
    if (!onClick) {
        classes += " disabled"
    }

    if (currentIndex) {
        classes += " active-pagination-btn"
    }

    return(
        <FlexBox className={classes} onClick={onClick}>
            <div style={{textAlign: "center"}}>
                {label}
            </div>
        </FlexBox>
    )
}