import { Boxes } from "lucide-react";
import { useTranslation } from "react-i18next";

const NoResult = () => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col items-center justify-center gap-1 p-10">
      <Boxes />
      <span className="text-center text-sm">
        {t("pages.instances.list.empty.title")}
      </span>
    </div>
  );
};

export default NoResult;
