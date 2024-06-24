import { Dialog, DialogContent } from "~/design/Dialog";
import {
  EventsPageSizeValueSchema,
  eventsPageSizeValue,
  useEventsPageSize,
  useEventsPageSizeActions,
} from "~/util/store/events";
import { NoPermissions, NoResult, TableCell, TableRow } from "~/design/Table";
import { Pagination, PaginationLink } from "~/design/Pagination";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

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
  const { t } = useTranslation();
  const [eventDialog, setEventDialog] = useState<EventSchemaType | null>();

  const { setEventsPageSize } = useEventsPageSizeActions();
  const pageSize = useEventsPageSize();

  const { data, isFetched, isAllowed, noPermissionMessage } = useEvents({
    enabled: true,
    filters,
  });

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
        <PaginationProvider items={data.data} pageSize={parseInt(pageSize)}>
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
                      const parseValue =
                        EventsPageSizeValueSchema.safeParse(value);
                      if (parseValue.success) {
                        setEventsPageSize(parseValue.data);
                        goToPage(1);
                      }
                    }}
                  >
                    <SelectTrigger variant="outline">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {eventsPageSizeValue.map((size) => (
                        <SelectItem key={size} value={size}>
                          {/* TODO: i18n */}
                          Show {size} rows
                        </SelectItem>
                      ))}
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
