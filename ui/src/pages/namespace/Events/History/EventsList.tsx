import { Dialog, DialogContent } from "~/design/Dialog";
import { NoPermissions, NoResult, TableCell, TableRow } from "~/design/Table";
import { useEventsPageSize, usePageSizeActions } from "~/util/store/pagesize";

import { Card } from "~/design/Card";
import { EventSchemaType } from "~/api/events/schema";
import EventsTable from "./Table";
import Filters from "./components/Filters";
import { FiltersSchemaType } from "~/api/events/schema/filters";
import { Pagination } from "~/components/Pagination";
import PaginationProvider from "~/components/PaginationProvider";
import { Radio } from "lucide-react";
import Row from "./Row";
import { SelectPageSize } from "../../../../components/SelectPageSize";
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
  const pageSize = useEventsPageSize();
  const { setEventsPageSize } = usePageSizeActions();
  const [eventDialog, setEventDialog] = useState<EventSchemaType | null>();

  const { data, isFetched, isAllowed, noPermissionMessage } = useEvents({
    enabled: true,
    filters,
  });

  const { t } = useTranslation();

  const handleOpenChange = (state: boolean) => {
    if (!state) {
      setEventDialog(null);
    }
  };

  if (!data) return null;

  const noResults = isFetched && data.data.length === 0;
  const hasFilters = !!Object.keys(filters).length;

  const totalPages = Math.max(
    1,
    Math.ceil(data.data.length / Number(pageSize))
  );

  return (
    <div className="flex grow flex-col gap-y-3 p-5">
      <Dialog open={!!eventDialog} onOpenChange={handleOpenChange}>
        <PaginationProvider items={data.data} pageSize={parseInt(pageSize)}>
          {({
            currentItems,
            goToFirstPage,
            goToPage,
            // goToNextPage,
            // goToPreviousPage,
            currentPage,
            // pagesList,
          }) => (
            <>
              <Card>
                <div className="flex flex-row place-content-between items-start">
                  {currentPage}
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
              <div className="flex items-center justify-end gap-2">
                <SelectPageSize
                  initialPageSize={pageSize}
                  onSelect={(selectedSize) => {
                    setEventsPageSize(selectedSize);
                    goToFirstPage();
                  }}
                />
                <Pagination
                  totalPages={totalPages}
                  value={currentPage}
                  onChange={(value) => {
                    goToPage(value);
                  }}
                />
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
