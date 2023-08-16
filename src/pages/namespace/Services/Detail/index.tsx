import {
  Table,
  TableBody,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Card } from "~/design/Card";
import { Diamond } from "lucide-react";
import { pages } from "~/util/router/pages";
import { useServiceDetails } from "~/api/services/query/details";
import { useTranslation } from "react-i18next";

const ServiceDetailPage = () => {
  const { service } = pages.services.useParams();

  const { t } = useTranslation();

  const { data } = useServiceDetails({
    service: service ?? "",
  });
  if (!service) return null;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex flex-col gap-4 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
          <Diamond className="h-5" />
          {service}
        </h3>
      </div>
      <Card>
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
          <TableBody></TableBody>
        </Table>
      </Card>
      {data?.revisions?.map((revision) => (
        <div key={revision.name}>{revision.name}</div>
      ))}
    </div>
  );
};

export default ServiceDetailPage;
