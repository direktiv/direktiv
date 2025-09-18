import { VariableInput } from "../../VariableInput";
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
    <VariableInput
      placeholder={t("direktivPage.blockEditor.blockForms.keyValue.variable")}
      value={value}
      onUpdate={(value) => onChange(value)}
    />
  );
};
