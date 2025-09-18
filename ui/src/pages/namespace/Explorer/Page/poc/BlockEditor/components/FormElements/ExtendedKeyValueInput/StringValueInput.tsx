import Input from "~/design/Input";
import { SmartInput } from "../../SmartInput";
import { useTranslation } from "react-i18next";

type StringValueInputProps = {
  value: string;
  onChange: (value: string) => void;
  smart?: boolean;
};

export const StringValueInput = ({
  value,
  onChange,
  smart = false,
}: StringValueInputProps) => {
  const { t } = useTranslation();
  return smart ? (
    <SmartInput
      placeholder={t("direktivPage.blockEditor.blockForms.keyValue.value")}
      value={value}
      onUpdate={onChange}
    />
  ) : (
    <Input
      placeholder={t("direktivPage.blockEditor.blockForms.keyValue.value")}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  );
};
