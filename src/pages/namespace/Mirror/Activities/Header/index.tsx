import { GitCompare } from "lucide-react";
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
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.mirror.header.url")}
          </div>
          {repo}
        </div>
      </div>
    </div>
  );
};

export default Header;
