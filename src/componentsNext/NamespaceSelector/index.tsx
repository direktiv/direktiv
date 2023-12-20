import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { FC } from "react";
import { Loader2 } from "lucide-react";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

export type ButtonProps = {
  onValueChange?: (value: string) => void;
};

const NamespaceSelector: FC<ButtonProps> = ({ onValueChange }) => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  const {
    data: availableNamespaces,
    isLoading,
    isSuccess,
  } = useListNamespaces();

  if (!namespace) return null;

  const hasResults = isSuccess && availableNamespaces?.results.length > 0;

  return (
    <Select onValueChange={onValueChange}>
      <SelectTrigger variant="outline">
        <SelectValue
          placeholder={t("components.namespaceSelector.placeholder")}
        />
      </SelectTrigger>
      {isLoading && (
        <SelectContent>
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          {t("components.namespaceSelector.placeholder")}
        </SelectContent>
      )}
      {hasResults && (
        <SelectContent>
          {availableNamespaces?.results.map((ns) => (
            <SelectItem key={ns.name} value={ns.name}>
              <span>{ns.name}</span>
            </SelectItem>
          ))}
        </SelectContent>
      )}
    </Select>
  );
};
export default NamespaceSelector;
