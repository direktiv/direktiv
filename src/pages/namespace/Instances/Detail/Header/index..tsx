import { Box, FileSymlink } from "lucide-react";

import Badge from "~/design/Badge";
import Button from "~/design/Button";
import ChildInstances from "./ChildInstances";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { statusToBadgeVariant } from "../../utils";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "../store/instanceContext";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const Header = () => {
  const instanceId = useInstanceId();
  const { data } = useInstanceDetails({ instanceId });

  const { t } = useTranslation();

  const [invoker] = (data?.instance?.invoker ?? "").split(":");
  const updatedAt = useUpdatedAt(data?.instance.updatedAt);
  const createdAt = useUpdatedAt(data?.instance.createdAt);

  if (!data) return null;

  const link = pages.explorer.createHref({
    path: data.instance.as,
    namespace: data.namespace,
    subpage: "workflow",
  });

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-start">
        <div className="flex flex-col items-start gap-2">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <Box className="h-5" /> {data.instance.id.slice(0, 8)}
          </h3>
          <Badge
            variant={statusToBadgeVariant(data.instance.status)}
            icon={data.instance.status}
          >
            {data.instance.status}
          </Badge>
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.instances.detail.header.invoker")}
          </div>
          {invoker}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.instances.detail.header.startedAt")}
          </div>
          {t("pages.instances.detail.header.realtiveTime", {
            relativeTime: createdAt,
          })}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.instances.detail.header.updatedAt")}
          </div>
          {t("pages.instances.detail.header.realtiveTime", {
            relativeTime: updatedAt,
          })}
        </div>
        <ChildInstances />
        <div className="grow justify-end md:flex">
          <Button asChild variant="primary" className="max-md:w-full">
            <Link to={link}>
              <FileSymlink />
              {t("pages.instances.detail.header.openWorkflow")}
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
};

export default Header;
