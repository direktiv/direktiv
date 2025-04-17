import {
  PageSizeValueSchema,
  PageSizeValueType,
  pageSizeValue,
} from "~/util/store/pagesize";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { useTranslation } from "react-i18next";

export const SelectPageSize = ({
  onSelect,
  initialPageSize,
}: {
  onSelect: (selectedSize: PageSizeValueType) => void;
  initialPageSize: string;
}) => {
  const { t } = useTranslation();

  return (
    <Select
      defaultValue={initialPageSize}
      onValueChange={(value) => {
        const parseValue = PageSizeValueSchema.safeParse(value);
        if (parseValue.success) {
          onSelect(parseValue.data);
        }
      }}
    >
      <SelectTrigger variant="outline">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {pageSizeValue.map((size) => (
          <SelectItem key={size} value={size}>
            {t("pages.events.history.selectPageSize.selectItem", {
              count: Number(size),
            })}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
};
