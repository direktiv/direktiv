import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { mimeTypes } from "./utils";
import { useTranslation } from "react-i18next";

const MimeTypeSelect = ({
  id,
  mimeType,
  onChange,
  loading = false,
}: {
  id?: string;
  loading?: boolean;
  mimeType: string | undefined;
  onChange: (value: string) => void;
}) => {
  const { t } = useTranslation();
  const hasEditableMimeType = !!mimeType;
  const mimeTypeIsEmpty = mimeType === "";
  return (
    <Select
      onValueChange={onChange}
      defaultValue={mimeType}
      value={!hasEditableMimeType ? undefined : mimeType}
    >
      <SelectTrigger
        id={id}
        loading={loading}
        variant="outline"
        block
        disabled={!hasEditableMimeType}
      >
        <SelectValue
          placeholder={t("components.variableForm.mimeType.placeholder")}
        >
          {mimeTypeIsEmpty ? (
            <i>{t("components.variableForm.mimeType.empty")}</i>
          ) : (
            mimeType
          )}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        {mimeTypes.map((type) => (
          <SelectItem key={type.value} value={type.value}>
            {type.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
};

export default MimeTypeSelect;
