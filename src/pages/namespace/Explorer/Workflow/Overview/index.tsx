import {
  Boxes,
  Layers,
  Network,
  PieChart,
  RefreshCcw,
  RotateCcw,
} from "lucide-react";
import { Dialog, DialogContent } from "~/design/Dialog";
import { FC, useState } from "react";
import { NoResult, Table, TableBody } from "~/design/Table";
import {
  ServicesStreamingSubscriber,
  useServices,
} from "~/api/services/query/getAll";
import { Trans, useTranslation } from "react-i18next";

import { Card } from "~/design/Card";
import { CategoryBar } from "@tremor/react";
import Delete from "~/pages/namespace/Services/List/Delete";
import { InstanceCard } from "~/pages/namespace/Monitoring/Instances/InstanceCard";
import { InstanceRow } from "~/pages/namespace/Monitoring/Instances/Row";
import Metrics from "./Metrics";
import RefreshButton from "~/design/RefreshButton";
import { ScrollArea } from "~/design/ScrollArea";
import { ServiceSchemaType } from "~/api/services/schema/services";
import ServicesTable from "~/pages/namespace/Services/List/Table";
import { forceLeadingSlash } from "~/api/tree/utils";
import { pages } from "~/util/router/pages";
import { useInstances } from "~/api/instances/query/get";
import { useMetrics } from "~/api/tree/query/metrics";
import { useRouter } from "~/api/tree/query/router";

const ActiveWorkflowPage: FC = () => {
  const [deleteService, setDeleteService] = useState<ServiceSchemaType>();
  const [dialogOpen, setDialogOpen] = useState(false);

  const { path } = pages.explorer.useParams();
  const { data: routerData } = useRouter({ path });
  const {
    data: instances,
    isFetching: isFetchingInstances,
    refetch: refetchInstances,
  } = useInstances({
    limit: 10,
    offset: 0,
    filters: { AS: { type: "WORKFLOW", value: forceLeadingSlash(path) } },
  });
  const {
    data: successData,
    isFetching: isFetchingSuccessful,
    refetch: refetchSuccessful,
  } = useMetrics({
    path,
    type: "successful",
  });
  const {
    data: failedData,
    isFetching: isFetchingFailed,
    refetch: refetchFailed,
  } = useMetrics({
    path,
    type: "failed",
  });
  const { data: servicesData, isSuccess: servicesIsSuccess } = useServices({
    workflow: path,
  });
  const { t } = useTranslation();

  const routes = routerData?.routes;

  const successful = Number(successData?.results[0]?.value[1]);
  const failed = Number(failedData?.results[0]?.value[1]);
  const metrics =
    successful || failed
      ? {
          successful: successful || 0,
          failed: failed || 0,
        }
      : undefined;

  const isFetchingMetrics = isFetchingFailed || isFetchingSuccessful;

  const refetchMetrics = () => {
    refetchSuccessful();
    refetchFailed();
  };

  const instancesRefetchButton = (
    <RefreshButton
      icon
      size="sm"
      variant="ghost"
      disabled={isFetchingInstances}
      onClick={() => {
        refetchInstances();
      }}
    />
  );

  const MetricsRefetchButton = () => (
    <RefreshButton
      icon
      size="sm"
      variant="ghost"
      disabled={isFetchingMetrics}
      onClick={() => {
        refetchMetrics();
      }}
    />
  );

  const DeleteMenuItem = () => {
    const { t } = useTranslation();
    return (
      <>
        <RefreshCcw className="mr-2 h-4 w-4" />
        {t("pages.explorer.tree.workflow.overview.services.deleteMenuItem")}
      </>
    );
  };

  return (
    <div className="grid gap-5 p-4 md:grid-cols-[2fr_1fr]">
      <InstanceCard
        className="row-span-2"
        headline={t("pages.explorer.tree.workflow.overview.instances.header")}
        icon={Boxes}
        refetchButton={instancesRefetchButton}
      >
        {instances?.instances?.results.length === 0 ? (
          <NoResult icon={Boxes}>
            {t("pages.explorer.tree.workflow.overview.instances.noResult")}
          </NoResult>
        ) : (
          <ScrollArea className="h-full">
            <Table>
              <TableBody>
                {instances?.instances?.results.map((instance) => (
                  <InstanceRow key={instance.id} instance={instance} />
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        )}
      </InstanceCard>

      <Card className="flex flex-col">
        <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
          <PieChart className="h-5" />
          <h3 className="grow">
            {t("pages.explorer.tree.workflow.overview.metrics.header")}
          </h3>
          <MetricsRefetchButton />
        </div>
        {metrics ? (
          <Metrics data={metrics} />
        ) : (
          <NoResult icon={PieChart}>
            {t("pages.explorer.tree.workflow.overview.metrics.noResult")}
          </NoResult>
        )}
      </Card>

      <Card>
        <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
          <Network className="h-5" />
          <h3 className="grow">
            {t(
              "pages.explorer.tree.workflow.overview.trafficDistribution.header"
            )}
          </h3>
        </div>
        {routes && routes[0] && routes[1] ? (
          <div className="p-5 pt-1">
            <CategoryBar
              values={[35, 65]}
              colors={["indigo", "gray"]}
              markerValue={35}
              className="mt-3"
            />
            <div className="flex flex-row justify-between">
              <div>{routes[0].ref.slice(0, 8)}</div>
              <div>{routes[1].ref.slice(0, 8)}</div>
            </div>
          </div>
        ) : (
          <NoResult>
            {t(
              "pages.explorer.tree.workflow.overview.trafficDistribution.noResult"
            )}
          </NoResult>
        )}
      </Card>

      <Card className="col-span-2">
        <ServicesStreamingSubscriber workflow={path} />
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
            <Layers className="h-5" />
            <h3 className="grow">
              {t("pages.explorer.tree.workflow.overview.services.header")}
            </h3>
          </div>

          <ServicesTable
            items={servicesData}
            isSuccess={servicesIsSuccess}
            setDeleteService={setDeleteService}
            deleteMenuItem={<DeleteMenuItem />}
          />

          <DialogContent>
            {deleteService && (
              <Delete
                icon={RotateCcw}
                header={t(
                  "pages.explorer.tree.workflow.overview.services.delete.title"
                )}
                message={
                  <Trans
                    i18nKey="pages.explorer.tree.workflow.overview.services.delete.message"
                    values={{ name: deleteService.info.name }}
                  />
                }
                service={deleteService.info.name}
                workflow={path}
                version={deleteService.info.revision}
                close={() => {
                  setDialogOpen(false);
                }}
              />
            )}
          </DialogContent>
        </Dialog>
      </Card>
    </div>
  );
};

export default ActiveWorkflowPage;
