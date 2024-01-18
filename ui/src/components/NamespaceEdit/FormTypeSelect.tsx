import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { MirrorFormType } from "~/api/tree/schema/mirror";
import { useTranslation } from "react-i18next";

const FormTypeSelect = ({
  value,
  isNew,
  storedValue,
  onValueChange,
}: {
  value: MirrorFormType;
  isNew: boolean;
  storedValue?: MirrorFormType;
  onValueChange: (value: MirrorFormType) => void;
}) => {
  const { t } = useTranslation();

  return (
    <Select value={value} onValueChange={onValueChange}>
      <SelectTrigger variant="outline" className="w-full">
        <SelectValue
          placeholder={t("components.namespaceEdit.placeholder.formType")}
        />
      </SelectTrigger>
      <SelectContent>
        {!isNew && storedValue === "keep-token" && (
          <SelectItem
            value="keep-token"
            onClick={() => onValueChange("keep-token")}
          >
            {t("components.namespaceEdit.formTypeSelect.keep-token")}
          </SelectItem>
        )}
        {!isNew && storedValue === "keep-ssh" && (
          <SelectItem
            value="keep-ssh"
            onClick={() => onValueChange("keep-ssh")}
          >
            {t("components.namespaceEdit.formTypeSelect.keep-ssh")}
          </SelectItem>
        )}
        <SelectItem value="public" onClick={() => onValueChange("public")}>
          {t("components.namespaceEdit.formTypeSelect.public")}
        </SelectItem>
        <SelectItem value="token" onClick={() => onValueChange("token")}>
          {t("components.namespaceEdit.formTypeSelect.token")}
        </SelectItem>
        <SelectItem value="ssh" onClick={() => onValueChange("ssh")}>
          {t("components.namespaceEdit.formTypeSelect.ssh")}
        </SelectItem>
      </SelectContent>
    </Select>
  );
};

export default FormTypeSelect;
