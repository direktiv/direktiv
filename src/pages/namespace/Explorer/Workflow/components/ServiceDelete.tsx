import { RefreshCcw, RotateCcw } from "lucide-react";
import { Trans, useTranslation } from "react-i18next";

import Delete from "~/pages/namespace/Services/List/Delete";
import { ServiceSchemaType } from "~/api/services/schema/services";

export const DeleteMenuItem = () => {
  const { t } = useTranslation();
  return (
    <>
      <RefreshCcw className="mr-2 h-4 w-4" />
      {t("pages.explorer.tree.workflow.overview.services.deleteMenuItem")}
    </>
  );
};

export const ServiceDelete = ({
  service,
  workflow,
  onClose,
}: {
  service: ServiceSchemaType;
  workflow: string;
  onClose: () => void;
}) => {
  const { t } = useTranslation();
  return (
    <Delete
      icon={RotateCcw}
      header={t("pages.explorer.tree.workflow.overview.services.delete.title")}
      message={
        <Trans
          i18nKey="pages.explorer.tree.workflow.overview.services.delete.message"
          values={{ name: service.info.name }}
        />
      }
      service={service.info.name}
      workflow={workflow}
      version={service.info.revision}
      close={() => {
        onClose();
      }}
    />
  );
};
