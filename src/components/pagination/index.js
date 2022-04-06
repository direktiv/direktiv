import React, { useCallback, useState } from 'react';
import './style.css';
import FlexBox from '../flexbox';
import {BsChevronBarLeft, BsChevronBarRight, BsChevronLeft, BsChevronRight} from 'react-icons/bs'



function Pagination(props) {

    const { queryParams, pageInfo, updatePage, pageSize=5, total=1 } = props;
    const [isFirstPage, setIsFirstPage] = useState(true)
    const [isLastPage, setIsLastPage] = useState(false)
    const defaultQuery = [`first=${pageSize}`]

    const handlePageChange = useCallback((direction)=>{
        let firstPage = false
        let lastPage = false
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
                firstPage = true
                break;
            case 'last':
                const rest = total%pageSize
                updatePage([`last=${(rest === 0)? pageSize: rest}`])
                lastPage = true
                break;
            default:   
                return    
        }
        setIsFirstPage(firstPage)
        setIsLastPage(lastPage)
    }, [pageInfo, total, pageSize, updatePage])
    
    let hasNext = pageInfo?.hasNextPage  && !isLastPage
    let hasPrev = (pageInfo?.hasPreviousPage && !isFirstPage) && (!queryParams || queryParams.toString() !== defaultQuery.toString())

    const hasNextClass = hasNext ? 'arrow active': 'arrow'
    const hasPrevClass = hasPrev ? 'arrow active': 'arrow'

    return(
        <FlexBox style={{justifyContent: "flex-end"}}>
        <FlexBox className="pagination-container" style={{}}>
            <FlexBox className={'pagination-btn'} style={{ maxWidth: "24px" }} onClick={() => {
                handlePageChange('first')
            }}>
                <BsChevronBarLeft className={'arrow active'} />
            </FlexBox>
            
            <FlexBox className={`pagination-btn ${hasPrev ? "":"disabled"}`} style={{ maxWidth: "24px" }} onClick={() => {
                handlePageChange('prev')
            }}>
                <BsChevronLeft className={hasPrevClass} />
            </FlexBox>
            <FlexBox style={{width: "40px"}}>

            </FlexBox>
            <FlexBox className={`pagination-btn ${hasNext ? "":"disabled"}`} style={{ maxWidth: "24px" }} onClick={() => {
                handlePageChange('next')
            }}>
                <BsChevronRight className={hasNextClass} />
            </FlexBox>

            <FlexBox className={`pagination-btn `} onClick={() => {
                handlePageChange('last')
            }}>
                <BsChevronBarRight className={'arrow active'} />
            </FlexBox>
        </FlexBox>
        </FlexBox>
    )
}

export default Pagination;