import { Dialog, DialogContent } from "~/design/Dialog";
import {
  NoPermissions,
  NoResult,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import { Pagination, PaginationLink } from "~/design/Pagination";

import { Card } from "~/design/Card";
import { EventSchemaType } from "~/api/eventsv2/schema";
import Filters from "./components/Filters";
import { FiltersObj } from "~/api/events/query/get";
import PaginationProvider from "~/components/PaginationProvider";
import { Radio } from "lucide-react";
import Row from "./Row";
import SendEvent from "./SendEvent";
import ViewEvent from "./ViewEvent";
import { useEvents } from "~/api/eventsv2/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const pageSize = 10;

const EventsList = ({
  filters,
  setFilters,
}: {
  filters: FiltersObj;
  setFilters: (filters: FiltersObj) => void;
}) => {
  const { t } = useTranslation();
  const [eventDialog, setEventDialog] = useState<EventSchemaType | null>();

  const { data, isFetched, isAllowed, noPermissionMessage } = useEvents({
    enabled: true,
  });

  const handleOpenChange = (state: boolean) => {
    if (!state) {
      setEventDialog(null);
    }
  };

  if (!data) return null;

  const noResults = isFetched && data.length === 0;
  const hasFilters = !!Object.keys(filters).length;

  return (
    <div className="flex grow flex-col gap-y-3 p-5">
      <Dialog open={!!eventDialog} onOpenChange={handleOpenChange}>
        <PaginationProvider items={data} pageSize={pageSize}>
          {({
            currentItems,
            goToFirstPage,
            goToPage,
            goToNextPage,
            goToPreviousPage,
            currentPage,
            pagesList,
            totalPages,
          }) => (
            <>
              <Card>
                <div className="flex flex-row place-content-between items-start">
                  <Filters
                    filters={filters}
                    onUpdate={(filters) => {
                      setFilters(filters);
                      goToFirstPage();
                    }}
                  />
                  <div className="m-2 flex flex-row flex-wrap gap-2">
                    <SendEvent />
                  </div>
                </div>

                <Table className="border-t border-gray-5 dark:border-gray-dark-5">
                  <TableHead>
                    <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                      <TableHeaderCell id="event-type">
                        {t("pages.events.history.tableHeader.type")}
                      </TableHeaderCell>
                      <TableHeaderCell id="event-id">
                        {t("pages.events.history.tableHeader.id")}
                      </TableHeaderCell>
                      <TableHeaderCell id="event-source">
                        {t("pages.events.history.tableHeader.source")}
                      </TableHeaderCell>
                      <TableHeaderCell id="event-received-at">
                        {t("pages.events.history.tableHeader.receivedAt")}
                      </TableHeaderCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {isAllowed ? (
                      <>
                        {noResults ? (
                          <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                            <TableCell colSpan={6}>
                              <NoResult icon={Radio}>
                                {hasFilters
                                  ? t(
                                      "pages.events.history.empty.noFilterResults"
                                    )
                                  : t("pages.events.history.empty.noResults")}
                              </NoResult>
                            </TableCell>
                          </TableRow>
                        ) : (
                          currentItems?.map((event) => (
                            <Row
                              key={event.event.id}
                              event={event.event}
                              receivedAt={event.receivedAt}
                              namespace={event.namespace}
                              onClick={setEventDialog}
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
                  </TableBody>
                </Table>
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

        <DialogContent className="sm:max-w-xl md:max-w-2xl lg:max-w-3xl">
          {!!eventDialog && (
            <ViewEvent
              event={eventDialog}
              handleOpenChange={handleOpenChange}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default EventsList;
