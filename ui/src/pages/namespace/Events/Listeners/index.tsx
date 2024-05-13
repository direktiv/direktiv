import { NoPermissions, NoResult, TableCell, TableRow } from "~/design/Table";
import { Pagination, PaginationLink } from "~/design/Pagination";

import { Antenna } from "lucide-react";
import { Card } from "~/design/Card";
import ListenersTable from "./Table";
import PaginationProvider from "~/components/PaginationProvider";
import Row from "./Row";
import { useEventListeners } from "~/api/eventListenersv2/query/get";
import { useTranslation } from "react-i18next";

const pageSize = 10;

const ListenersList = () => {
  const { data, isFetched, isAllowed, noPermissionMessage } = useEventListeners(
    { limit: 10, offset: 0 }
  );

  const { t } = useTranslation();

  if (!data?.data) return null;

  const noResults = isFetched && data.data.length === 0;

  return (
    <div className="flex grow flex-col gap-y-3 p-5">
      <PaginationProvider items={data.data} pageSize={pageSize}>
        {({
          currentItems,
          goToPage,
          goToNextPage,
          goToPreviousPage,
          currentPage,
          pagesList,
          totalPages,
        }) => (
          <>
            <Card>
              <ListenersTable>
                {isAllowed ? (
                  <>
                    {noResults ? (
                      <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                        <TableCell colSpan={6}>
                          <NoResult icon={Antenna}>
                            {t("pages.events.listeners.empty.noResults")}
                          </NoResult>
                        </TableCell>
                      </TableRow>
                    ) : (
                      currentItems.map((listener, i) => (
                        <Row
                          listener={listener}
                          key={i}
                          namespace={listener.namespace}
                          data-testid={`listener-row-${i}`}
                        />
                      ))
                    )}
                  </>
                ) : (
                  <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                    <TableCell colSpan={6}>
                      <NoPermissions>{noPermissionMessage}</NoPermissions>
                    </TableCell>
                  </TableRow>
                )}
              </ListenersTable>
            </Card>

            {totalPages > 1 && (
              <Pagination>
                <PaginationLink
                  data-testid="pagination-btn-left"
                  icon="left"
                  onClick={() => goToPreviousPage()}
                />
                {pagesList.map((page) => (
                  <PaginationLink
                    active={currentPage === page}
                    key={`${page}`}
                    onClick={() => goToPage(page)}
                  >
                    {page}
                  </PaginationLink>
                ))}
                <PaginationLink
                  data-testid="pagination-btn-right"
                  icon="right"
                  onClick={() => goToNextPage()}
                />
              </Pagination>
            )}
          </>
        )}
      </PaginationProvider>
    </div>
  );
};

export default ListenersList;
