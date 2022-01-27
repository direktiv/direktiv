import React, { useCallback } from 'react';
import './style.css';
import FlexBox from '../flexbox';
import {BsChevronBarLeft, BsChevronBarRight, BsChevronLeft, BsChevronRight} from 'react-icons/bs'



function Pagination(props) {

    const { pageInfo, updatePage, pageSize=5, total=1 } = props;

    const handlePageChange = useCallback((direction)=>{
        switch(direction){
            case 'next':
                if(pageInfo?.hasNextPage){
                    const after = `after=${pageInfo?.endCursor}`
                    const first = `first=${pageSize}`
                    updatePage([first, after])
                }
                break
            case 'prev':
                if(pageInfo?.hasPreviousPage){
                    const before = `before=${pageInfo?.startCursor}`
                    const first = `last=${pageSize}`
                    updatePage([before, first])
                }
                break
            case 'first': 
                updatePage([`first=${pageSize}`])
                break;
            case 'last':
                const rest = total%pageSize
                updatePage([`last=${(rest === 0)? pageSize: rest}`])
                break;
            default:   
                return    
        }
    }, [pageInfo, total, pageSize, updatePage])
    
    const hasNext = pageInfo?.hasNextPage? 'arrow active': 'arrow'
    const hasPrev = pageInfo?.hasPreviousPage? 'arrow active': 'arrow'

    return(
        <FlexBox style={{justifyContent: "flex-end"}}>
        <FlexBox className="pagination-container" style={{}}>
            <FlexBox className={'pagination-btn'} style={{ maxWidth: "24px" }} onClick={() => {
                handlePageChange('first')
            }}>
                <BsChevronBarLeft className={'arrow active'} />
            </FlexBox>
            
            <FlexBox className={'pagination-btn'} style={{ maxWidth: "24px" }} onClick={() => {
                handlePageChange('prev')
            }}>
                <BsChevronLeft className={hasPrev} />
            </FlexBox>
            <FlexBox style={{width: "40px"}}>

            </FlexBox>
            <FlexBox className={'pagination-btn'} style={{ maxWidth: "24px" }} onClick={() => {
                handlePageChange('next')
            }}>
                <BsChevronRight className={hasNext} />
            </FlexBox>

            <FlexBox className={'pagination-btn'} onClick={() => {
                handlePageChange('last')
            }}>
                <BsChevronBarRight className={'arrow active'} />
            </FlexBox>
        </FlexBox>
        </FlexBox>
    )
}

export default Pagination;