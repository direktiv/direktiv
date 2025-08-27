import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type VariableValueInputProps = {
  value: string;
  onChange: (value: string) => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
};

export const VariableValueInput = ({
  value,
  onChange,
  onKeyDown,
}: VariableValueInputProps) => {
  const { t } = useTranslation();
  return (
    <Input
      placeholder={t("direktivPage.blockEditor.blockForms.keyValue.variable")}
      value={value}
      onKeyDown={onKeyDown}
      onChange={(e) => onChange(e.target.value)}
    />
  );
};
