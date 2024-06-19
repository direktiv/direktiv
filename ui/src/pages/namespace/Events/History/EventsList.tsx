import { Dialog, DialogContent } from "~/design/Dialog";
import { NoPermissions, NoResult, TableCell, TableRow } from "~/design/Table";
import { Pagination, PaginationLink } from "~/design/Pagination";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import {
  useEventsPageSizeActions,
  useEventsPageSizeState,
} from "~/util/store/events";

import { Card } from "~/design/Card";
import { EventSchemaType } from "~/api/events/schema";
import EventsTable from "./Table";
import Filters from "./components/Filters";
import { FiltersSchemaType } from "~/api/events/schema/filters";
import PaginationProvider from "~/components/PaginationProvider";
import { Radio } from "lucide-react";
import Row from "./Row";
import SendEvent from "./SendEvent";
import ViewEvent from "./ViewEvent";
import { useEvents } from "~/api/events/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const EventsList = ({
  filters,
  setFilters,
}: {
  filters: FiltersSchemaType;
  setFilters: (filters: FiltersSchemaType) => void;
}) => {
  const { setEventsPageSize } = useEventsPageSizeActions();
  const { t } = useTranslation();
  const [eventDialog, setEventDialog] = useState<EventSchemaType | null>();

  const pagesize = useEventsPageSizeState();
  const [pageSize, setPageSize] = useState(
    pagesize.pagesize ? Number(pagesize.pagesize) : 10
  );

  const { data, isFetched, isAllowed, noPermissionMessage } = useEvents({
    enabled: true,
    filters,
  });

  type EventsPageSizeValueType = 10 | 20 | 30 | 50 | null;

  const handleOpenChange = (state: boolean) => {
    if (!state) {
      setEventDialog(null);
    }
  };

  if (!data) return null;

  const noResults = isFetched && data.data.length === 0;
  const hasFilters = !!Object.keys(filters).length;

  return (
    <div className="flex grow flex-col gap-y-3 p-5">
      <Dialog open={!!eventDialog} onOpenChange={handleOpenChange}>
        <PaginationProvider items={data.data} pageSize={pageSize}>
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

                <EventsTable>
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
                </EventsTable>
              </Card>
              <div className="flex items-center justify-end">
                <div className="m-2">
                  <Select
                    defaultValue={String(pageSize)}
                    onValueChange={(value) => {
                      setPageSize(Number(value));
                      setEventsPageSize(
                        Number(value) as EventsPageSizeValueType
                      );
                      goToPage(1);
                    }}
                  >
                    <SelectTrigger variant="outline">
                      <SelectValue placeholder="Show 10 rows" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="10">Show 10 rows</SelectItem>
                      <SelectItem value="20">Show 20 rows</SelectItem>
                      <SelectItem value="30">Show 30 rows</SelectItem>
                      <SelectItem value="50">Show 50 rows</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
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
              </div>
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
