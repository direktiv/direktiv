import { TableCell, TableRow } from "~/design/Table";

import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

export const SearchBar = ({
  query,
  onChange,
}: {
  query: string;
  onChange: (newValue: string) => void;
}) => {
  const { t } = useTranslation();
  return (
    <TableRow className="hover:bg-white/75">
      <TableCell colSpan={2}>
        <Input
          data-testid="queryField"
          className="sm:w-60"
          value={query}
          onChange={(e) => {
            onChange(e.target.value);
          }}
          placeholder={t("pages.explorer.tree.list.filter")}
        />
      </TableCell>
    </TableRow>
  );
};
