import { Folder, FolderOpen, Play } from "lucide-react";

import Button from "../../../../design/Button";
import { FC } from "react";
import { useTranslation } from "react-i18next";

const NoResult: FC = () => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col items-center gap-y-5 p-10">
      <div className="flex flex-col items-center justify-center gap-1">
        <FolderOpen />
        <span className="text-center text-sm">
          {t("pages.explorer.tree.list.empty.title")}
        </span>
      </div>
      <div className="flex flex-col gap-5 sm:flex-row">
        <Button>
          <Play />
          {t("pages.explorer.tree.list.empty.createWorkflow")}
        </Button>
        <Button variant="outline">
          <Folder />
          {t("pages.explorer.tree.list.empty.createDirectory")}
        </Button>
      </div>
    </div>
  );
};

export default NoResult;
