import { Dispatch, SetStateAction } from "react";
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

import { Layers } from "lucide-react";
import Row from "./Row";
import { ServiceSchemaType } from "~/api/services/schema/services";
import { useTranslation } from "react-i18next";

const ServicesTable = ({
  services,
  isSuccess = false,
  setRebuildService,
  isAllowed,
  noPermissionMessage,
}: {
  services?: ServiceSchemaType[];
  isSuccess: boolean;
  setRebuildService: Dispatch<SetStateAction<ServiceSchemaType | undefined>>;
  isAllowed: boolean;
  noPermissionMessage?: string;
}) => {
  const { t } = useTranslation();

  const showTable = (services?.length ?? 0) > 0;
  const noResults = isSuccess && services?.length === 0;

  return (
    <Table>
      <TableHead>
        <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
          <TableHeaderCell>
            {t("pages.services.list.tableHeader.name")}
          </TableHeaderCell>
          <TableHeaderCell className="w-48">
            {t("pages.services.list.tableHeader.image")}
          </TableHeaderCell>
          <TableHeaderCell className="w-16">
            {t("pages.services.list.tableHeader.scale")}
          </TableHeaderCell>
          <TableHeaderCell className="w-20">
            {t("pages.services.list.tableHeader.size")}
          </TableHeaderCell>
          <TableHeaderCell className="w-48">
            {t("pages.services.list.tableHeader.cmd")}
          </TableHeaderCell>
          <TableHeaderCell className="w-16" />
        </TableRow>
      </TableHead>
      <TableBody>
        {isAllowed ? (
          <>
            {showTable &&
              services?.map((service) => (
                <Row
                  service={service}
                  key={service.id}
                  setRebuildService={setRebuildService}
                />
              ))}
            {noResults && (
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableCell colSpan={6}>
                  <NoResult icon={Layers}>
                    {t("pages.services.list.empty.title")}
                  </NoResult>
                </TableCell>
              </TableRow>
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
  );
};

export default ServicesTable;
