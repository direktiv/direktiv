import { Diamond } from "lucide-react";
import { StatusBadge } from "../components/StatusBadge";
import { useService } from "~/api/services/query/getAll";
import { useTranslation } from "react-i18next";

const Header = ({ service }: { service: string }) => {
  const { data: serviceData } = useService(service);
  const { t } = useTranslation();

  if (!serviceData) return null;

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-3 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 font-bold text-primary-500">
          <Diamond className="h-5" />
          {serviceData.name}
        </h3>
        {/* TODO: may link to path */}
        <div>{serviceData.filePath}</div>
      </div>
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-between">
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.list.tableHeader.image")}
          </div>
          {serviceData.image ? serviceData.image : "-"}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.list.tableHeader.scale")}
          </div>
          {serviceData.scale}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.list.tableHeader.size")}
          </div>
          {serviceData.size ? serviceData.size : "-"}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.list.tableHeader.cmd")}
          </div>
          {serviceData.cmd ? serviceData.cmd : "-"}
        </div>
      </div>
      <div>
        <div className="flex flex-col gap-3 sm:flex-row">
          {serviceData.error && (
            <StatusBadge
              status="False"
              className="w-fit"
              message={serviceData.error}
            >
              {t("pages.services.list.tableRow.errorLabel")}
            </StatusBadge>
          )}
          {(serviceData.conditions ?? []).map((condition) => (
            <StatusBadge
              key={condition.type}
              status={condition.status}
              message={condition.message}
              className="self-start"
            >
              {condition.type}
            </StatusBadge>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Header;
