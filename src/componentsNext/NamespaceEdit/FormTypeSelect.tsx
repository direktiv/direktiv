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
        <SelectValue placeholder="TODO: ADD PLACEHOLDER" />
      </SelectTrigger>
      <SelectContent>
        {!isNew && storedValue === "keep-token" && (
          <SelectItem
            value="keep-token"
            onClick={() => onValueChange("keep-token")}
          >
            Keep existing token
          </SelectItem>
        )}
        {!isNew && storedValue === "keep-ssh" && (
          <SelectItem
            value="keep-ssh"
            onClick={() => onValueChange("keep-ssh")}
          >
            Keep existing SSH keys
          </SelectItem>
        )}
        <SelectItem value="public" onClick={() => onValueChange("public")}>
          Public
        </SelectItem>
        <SelectItem value="token" onClick={() => onValueChange("token")}>
          Token
        </SelectItem>
        <SelectItem value="ssh" onClick={() => onValueChange("ssh")}>
          SSH
        </SelectItem>
      </SelectContent>
    </Select>
  );
};

export default FormTypeSelect;
