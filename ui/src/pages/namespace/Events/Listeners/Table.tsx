import { FC, PropsWithChildren } from "react";
import {
  Table,
  TableBody,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { useTranslation } from "react-i18next";

const ListenersTable: FC<PropsWithChildren> = ({ children }) => {
  const { t } = useTranslation();

  return (
    <Table className="border-t border-gray-5 dark:border-gray-dark-5">
      <TableHead>
        <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
          <TableHeaderCell>
            {t("pages.events.listeners.tableHeader.type")}
          </TableHeaderCell>
          <TableHeaderCell>
            {t("pages.events.listeners.tableHeader.target")}
          </TableHeaderCell>
          <TableHeaderCell>
            {t("pages.events.listeners.tableHeader.mode")}
          </TableHeaderCell>
          <TableHeaderCell>
            {t("pages.events.listeners.tableHeader.createdAt")}
          </TableHeaderCell>
          <TableHeaderCell>
            {t("pages.events.listeners.tableHeader.eventTypes")}
          </TableHeaderCell>
        </TableRow>
      </TableHead>
      <TableBody>{children}</TableBody>
    </Table>
  );
};

export default ListenersTable;
