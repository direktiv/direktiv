import { Dialog, DialogContent } from "~/design/Dialog";
import { Dispatch, SetStateAction, useState } from "react";
import { FiltersObj, useEvents } from "~/api/events/query/get";
import {
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
import NoResult from "./NoResult";
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

  const { data, isFetched } = useEvents({
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
      <div className="flex flex-row justify-between">
        <h3 className="flex items-center gap-x-2 pb-2 pt-1 font-bold">
          <Radio className="h-5" />
          {t("pages.events.history.title")} {eventDialog ? "open" : "close"}
        </h3>

        <SendEvent />
      </div>

      {/* <div className="flex flex-row justify-between">
        <h3 className="flex items-center gap-x-2 pt-1 font-bold"></h3>
      </div> */}

      <Dialog open={!!eventDialog} onOpenChange={handleOpenChange}>
        <Card>
          <Filters filters={filters} onUpdate={handleFilterChange} />

          <Table className="border-t border-gray-5 dark:border-gray-dark-5">
            <TableHead>
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableHeaderCell>
                  {t("pages.events.history.tableHeader.type")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.events.history.tableHeader.id")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.events.history.tableHeader.source")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.events.history.tableHeader.receivedAt")}
                </TableHeaderCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {noResults ? (
                <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                  <TableCell colSpan={6}>
                    <NoResult
                      message={
                        hasFilters
                          ? t("pages.events.history.empty.noFilterResults")
                          : t("pages.events.history.empty.noResults")
                      }
                    />
                  </TableCell>
                </TableRow>
              ) : (
                data?.events.results.map((event) => (
                  <Row
                    onClick={setEventDialog}
                    event={event}
                    key={event.id}
                    namespace={data.namespace}
                    data-testid={`event-row-${event.id}`}
                  />
                ))
              )}
            </TableBody>
          </Table>
        </Card>
        <DialogContent>
          {!!eventDialog && <ViewEvent event={eventDialog} />}
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
