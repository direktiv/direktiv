import { Box, FileSymlink } from "lucide-react";

import Button from "~/design/Button";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";

const Header = () => {
  const { t } = useTranslation();
  // TODO: update all fields
  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-start">
        <div className="flex flex-col items-start gap-2">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <Box className="h-5" /> 111
          </h3>
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.instances.detail.header.invoker")}
          </div>
          111
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.instances.detail.header.startedAt")}
          </div>
          dedede
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.instances.detail.header.updatedAt")}
          </div>
          dede
        </div>
        <div className="flex grow justify-end gap-5">
          <Button asChild isAnchor variant="primary" className="max-md:w-full">
            <Link to="">
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
