import { ComponentProps, FC } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { Loader2 } from "lucide-react";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

export type ButtonProps = ComponentProps<typeof SelectTrigger> & {
  defaultValue?: string;
  onValueChange?: (value: string) => void;
};

const NamespaceSelector: FC<ButtonProps> = ({
  defaultValue,
  onValueChange,
  ...props
}) => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  const { data: availableNamespaces, isLoading } = useListNamespaces();

  if (!namespace) return null;

  const defaultDoesNotExist =
    defaultValue &&
    !availableNamespaces?.results.some((ns) => ns.name === defaultValue);

  return (
    <Select onValueChange={onValueChange} defaultValue={defaultValue}>
      <SelectTrigger variant="outline" {...props}>
        <SelectValue
          placeholder={t("components.namespaceSelector.placeholder")}
        />
      </SelectTrigger>
      <SelectContent>
        {isLoading && (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            {t("components.namespaceSelector.placeholder")}
          </>
        )}
        {defaultDoesNotExist && (
          <SelectItem value={defaultValue}>
            <span>
              {t("components.namespaceSelector.optionDoesNotExist", {
                namespace: defaultValue,
              })}
            </span>
          </SelectItem>
        )}
        {availableNamespaces?.results.map((ns) => (
          <SelectItem key={ns.name} value={ns.name}>
            <span>{ns.name}</span>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
};
export default NamespaceSelector;
