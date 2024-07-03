import { TableCell, TableRow } from "~/design/Table";

import { FolderUp } from "lucide-react";
import { Link } from "react-router-dom";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

export const LevelUpNavigation = ({
  namespace,
  path,
}: {
  namespace: string;
  path?: string;
}) => {
  const pages = usePages();
  const { t } = useTranslation();
  return (
    <TableRow>
      <TableCell colSpan={2}>
        <Link
          to={pages.explorer.createHref({
            namespace,
            path,
          })}
          className="flex items-center space-x-3 hover:underline"
        >
          <FolderUp className="h-5" />
          <span>{t("pages.explorer.tree.list.oneLevelUp")}</span>
        </Link>
      </TableCell>
    </TableRow>
  );
};
