import Badge from "~/design/Badge";
import { FC } from "react";
import { useTranslation } from "react-i18next";

type AllowAnonymousProps = {
  allow: boolean;
};

export const AllowAnonymous: FC<AllowAnonymousProps> = ({ allow }) => {
  const { t } = useTranslation();
  return (
    <Badge
      icon={allow ? "complete" : "failed"}
      variant={allow ? "success" : "destructive"}
    >
      {allow
        ? t("pages.gateway.routes.row.allowAnonymous.yes")
        : t("pages.gateway.routes.row.allowAnonymous.no")}
    </Badge>
  );
};
