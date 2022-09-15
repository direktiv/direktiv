import { default as MUIPagination } from '@mui/material/Pagination';
import PaginationItem from '@mui/material/PaginationItem/PaginationItem';
import * as React from 'react';
export interface PageInfo {
  order: Order[]
  filter: Filter[]
  limit: number
  offset: number
  total: number
}

export interface Order {
  field: string
  direction: string
}

export interface Filter {
  field: string
  type: string
  val: string
}

export interface PageHandler {
  pageParams: string
  page: number
  pageCount: number
  offset: number
  limit: number
  updatePage: (newPage: number) => void
  goToFirstPage: () => void
}

/**
* Hook to track and update the current state of a pagination component. Additionally a goToFirstPage util function is returned. 
*/
export function usePageHandler(limit: number, initPage: number = 1): PageHandler {
  const [page, setPage] = React.useState(initPage)
  const offset = React.useMemo(() => {
    return (page - 1) * limit
  }, [page, limit])

  const pageParams = React.useMemo(() => {
    return `limit=${limit}&offset=${offset}`
  }, [offset, limit])

  const pageCount = 0

  const updatePage = React.useCallback((newPage: number) => {
    if (newPage !== page) {
      setPage(newPage)
    }
  }, [page, setPage])

  const goToFirstPage = React.useCallback(() => {
    setPage(1)
  }, [setPage])

  return {
    pageParams,
    page,
    offset,
    limit,
    pageCount,
    updatePage,
    goToFirstPage
  }
}


export interface PaginationProps {
  /**
  * Page Handler returned from the usePageHandler. 
  */
  pageHandler: PageHandler
  /**
  * Current page info of list edges.
  */
  pageInfo: PageInfo | null
}

/**
* A Pagaintion component that renders the lenght of items based on the passed pageInfo prop and its current state based on the pageHandler hook.
*/
function Pagination({ pageHandler, pageInfo }: PaginationProps) {
  const pageCount = React.useMemo(() => {
    if (!pageInfo || pageInfo.limit === 0) {
      return 0
    }

    return Math.ceil(pageInfo.total / pageInfo.limit)
  }, [pageInfo]);


  if (!pageInfo) {
    return (
      <div>
      </div>
    )
  }

  return (
    <MUIPagination
      renderItem={(item) => (
        <PaginationItem
          sx={{
            width: "fit-content",
            minWidth: "23px",
            fontWeight: "500",
            "&.Mui-selected": {
              boxShadow: "2px 2px 6px rgba(86, 104, 117, 0.16)"
            }
          }}
          {...item}
        />
      )}
      size="small"
      page={pageHandler.page}
      count={pageCount}
      color="primary"
      shape="rounded"
      onChange={(e, p) => {
        pageHandler.updatePage(p)
      }} />
  );
}

export default Pagination