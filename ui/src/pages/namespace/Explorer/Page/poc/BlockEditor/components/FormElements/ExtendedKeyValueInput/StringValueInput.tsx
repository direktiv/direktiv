import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type StringValueInputProps = {
  value: string;
  onChange: (value: string) => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
};

export const StringValueInput = ({
  value,
  onChange,
  onKeyDown,
}: StringValueInputProps) => {
  const { t } = useTranslation();
  return (
    <Input
      placeholder={t("direktivPage.blockEditor.blockForms.keyValue.value")}
      value={value}
      onKeyDown={onKeyDown}
      onChange={(e) => onChange(e.target.value)}
    />
  );
};
