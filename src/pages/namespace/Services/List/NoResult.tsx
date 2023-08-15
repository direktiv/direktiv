import { FC, PropsWithChildren } from "react";

import { Layers } from "lucide-react";
import { useTranslation } from "react-i18next";

const NoResult: FC<PropsWithChildren> = ({ children }) => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col items-center gap-y-5 p-10">
      <div className="flex flex-col items-center justify-center gap-1">
        <Layers />
        <span className="text-center text-sm">
          {t("pages.services.list.empty.title")}
        </span>
      </div>
      <div className="flex flex-col gap-5 sm:flex-row">{children}</div>
    </div>
  );
};

export default NoResult;
