import { FC, PropsWithChildren } from "react";
import {
  Table,
  TableBody,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { useTranslation } from "react-i18next";

const EventsTable: FC<PropsWithChildren> = ({ children }) => {
  const { t } = useTranslation();

  return (
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
      <TableBody>{children}</TableBody>
    </Table>
  );
};

export default EventsTable;
