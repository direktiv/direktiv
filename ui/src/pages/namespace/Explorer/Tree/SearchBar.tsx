import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { X } from "lucide-react";
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
        <div>
          <InputWithButton className="sm:w-60">
            <Input
              data-testid="queryField"
              value={query}
              onChange={(e) => {
                onChange(e.target.value);
              }}
              placeholder={t("pages.explorer.tree.list.filter")}
            />

            <Button
              icon
              variant="ghost"
              onClick={() => {
                onChange("");
              }}
            >
              <X />
            </Button>
          </InputWithButton>
        </div>
      </TableCell>
    </TableRow>
  );
};
