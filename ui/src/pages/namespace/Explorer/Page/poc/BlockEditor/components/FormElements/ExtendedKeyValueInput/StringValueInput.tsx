import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type StringValueInputProps = {
  value: string;
  onChange: (value: string) => void;
};

export const StringValueInput = ({
  value,
  onChange,
}: StringValueInputProps) => {
  const { t } = useTranslation();
  return (
    <Input
      placeholder={t("direktivPage.blockEditor.blockForms.keyValue.value")}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  );
};
