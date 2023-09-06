import { Dispatch, SetStateAction } from "react";
import {
  NoResult,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import {
  ServiceSchemaType,
  ServicesListSchemaType,
} from "~/api/services/schema/services";

import { Layers } from "lucide-react";
import Row from "./Row";
import { useTranslation } from "react-i18next";

const ServicesTable = ({
  items,
  isSuccess = false,
  setDeleteService,
  createNewButton,
  deleteMenuItem,
}: {
  items?: ServicesListSchemaType;
  isSuccess: boolean;
  setDeleteService: Dispatch<SetStateAction<ServiceSchemaType | undefined>>;
  createNewButton?: JSX.Element;
  deleteMenuItem?: JSX.Element;
}) => {
  const { t } = useTranslation();

  const showTable = (items?.functions?.length ?? 0) > 0;
  const noResults = isSuccess && items?.functions?.length === 0;

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
        {showTable &&
          items?.functions.map((service) => (
            <Row
              service={service}
              key={service.serviceName}
              setDeleteService={setDeleteService}
              deleteMenuItem={deleteMenuItem}
            />
          ))}
        {noResults && (
          <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
            <TableCell colSpan={6}>
              <NoResult icon={Layers} button={createNewButton}>
                {t("pages.services.list.empty.title")}
              </NoResult>
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  );
};

export default ServicesTable;
