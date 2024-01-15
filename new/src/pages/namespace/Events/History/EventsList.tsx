import { Dialog, DialogContent } from "~/design/Dialog";
import { Dispatch, SetStateAction, useState } from "react";
import { FiltersObj, useEvents } from "~/api/events/query/get";
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

import { Card } from "~/design/Card";
import { EventSchemaType } from "~/api/events/schema";
import Filters from "./components/Filters";
import { Pagination } from "~/componentsNext/Pagination";
import { Radio } from "lucide-react";
import Row from "./Row";
import SendEvent from "./SendEvent";
import ViewEvent from "./ViewEvent";
import { itemsPerPage } from ".";
import { useTranslation } from "react-i18next";

const EventsList = ({
  filters,
  setFilters,
  offset,
  setOffset,
}: {
  filters: FiltersObj;
  setFilters: (filters: FiltersObj) => void;
  offset: number;
  setOffset: Dispatch<SetStateAction<number>>;
}) => {
  const { t } = useTranslation();
  const [eventDialog, setEventDialog] = useState<EventSchemaType | null>();

  const { data, isFetched, isAllowed, noPermissionMessage } = useEvents({
    limit: itemsPerPage,
    offset,
    filters,
  });

  const handleOpenChange = (state: boolean) => {
    if (!state) {
      setEventDialog(null);
    }
  };

  const handleFilterChange = (filters: FiltersObj) => {
    setFilters(filters);
    setOffset(0);
  };

  const numberOfResults = data?.events?.pageInfo?.total ?? 0;
  const noResults = isFetched && data?.events.results.length === 0;
  const showPagination = numberOfResults > itemsPerPage;
  const hasFilters = !!Object.keys(filters).length;

  return (
    <div className="flex grow flex-col gap-y-3 p-5">
      <Dialog open={!!eventDialog} onOpenChange={handleOpenChange}>
        <Card>
          <div className="flex flex-row place-content-between items-start">
            <Filters filters={filters} onUpdate={handleFilterChange} />
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
                            ? t("pages.events.history.empty.noFilterResults")
                            : t("pages.events.history.empty.noResults")}
                        </NoResult>
                      </TableCell>
                    </TableRow>
                  ) : (
                    data?.events.results.map((event) => (
                      <Row
                        onClick={setEventDialog}
                        event={event}
                        key={event.id}
                        namespace={data.namespace}
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
        <DialogContent className="sm:max-w-xl md:max-w-2xl lg:max-w-3xl">
          {!!eventDialog && (
            <ViewEvent
              event={eventDialog}
              handleOpenChange={handleOpenChange}
            />
          )}
        </DialogContent>
      </Dialog>
      {showPagination && (
        <Pagination
          itemsPerPage={itemsPerPage}
          offset={offset}
          setOffset={(value) => setOffset(value)}
          totalItems={numberOfResults}
        />
      )}
    </div>
  );
};

export default EventsList;
