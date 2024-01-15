import { Diamond } from "lucide-react";
import { Link } from "react-router-dom";
import RefreshButton from "~/design/RefreshButton";
import { StatusBadge } from "../components/StatusBadge";
import { linkToServiceSource } from "../components/utils";
import { useService } from "~/api/services/query/services";
import { useTranslation } from "react-i18next";

const Header = ({ serviceId }: { serviceId: string }) => {
  const { data: service, refetch, isFetching } = useService(serviceId);
  const { t } = useTranslation();

  if (!service) return null;

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-3 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 font-bold text-primary-500">
          <Diamond className="h-5" />
          {service.name}
        </h3>

        <div>
          <Link className="hover:underline" to={linkToServiceSource(service)}>
            {service.filePath}
          </Link>
        </div>
      </div>
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-between">
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.list.tableHeader.image")}
          </div>
          {service.image ? service.image : "-"}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.list.tableHeader.scale")}
          </div>
          {service.scale}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.list.tableHeader.size")}
          </div>
          {service.size ? service.size : "-"}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.list.tableHeader.cmd")}
          </div>
          {service.cmd ? service.cmd : "-"}
        </div>
      </div>
      <div>
        <div className="flex flex-col items-center gap-3 sm:flex-row">
          {service.error && (
            <StatusBadge
              status="False"
              className="w-fit"
              message={service.error}
            >
              {t("pages.services.list.tableRow.errorLabel")}
            </StatusBadge>
          )}
          {(service.conditions ?? []).map((condition) => (
            <StatusBadge
              key={condition.type}
              status={condition.status}
              message={condition.message}
              className="self-start"
            >
              {condition.type}
            </StatusBadge>
          ))}
          <RefreshButton
            icon
            size="sm"
            variant="ghost"
            disabled={isFetching}
            onClick={() => {
              refetch();
            }}
          />
        </div>
      </div>
    </div>
  );
};

export default Header;
