import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type VariableValueInputProps = {
  value: string;
  onChange: (value: string) => void;
};

export const VariableValueInput = ({
  value,
  onChange,
}: VariableValueInputProps) => {
  const { t } = useTranslation();
  return (
    <Input
      placeholder={t("direktivPage.blockEditor.blockForms.keyValue.variable")}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  );
};
