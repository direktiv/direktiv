import { FileCog, GitCompare, RefreshCcw } from "lucide-react";

import Button from "~/design/Button";
import { useTranslation } from "react-i18next";

const Header = ({ name, repo }: { name: string; repo: string }) => {
  const { t } = useTranslation();

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-start">
        <div className="flex flex-col items-start gap-2">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <GitCompare className="h-5" /> {name}
          </h3>
          <div className="text-sm">{repo}</div>
        </div>
        <div className="flex grow justify-end gap-4">
          <Button variant="outline" className="max-md:w-full">
            <FileCog />
            {t("pages.mirror.header.editMirror")}
          </Button>
          <Button variant="primary" className="max-md:w-full">
            <RefreshCcw />
            {t("pages.mirror.header.sync")}
          </Button>
        </div>
      </div>
    </div>
  );
};

export default Header;
