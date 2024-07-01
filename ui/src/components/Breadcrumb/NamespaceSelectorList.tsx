import { Circle, Loader2 } from "lucide-react";
import {
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandStaticItem,
} from "~/design/Command";

import { Checkbox } from "~/design/Checkbox";
import { twMergeClsx } from "~/util/helpers";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

type NamespaceSelectorListProps =
  | {
      onSelectNamespace: (value: string) => void;
    } & (
      | { isMulti: false; selectedValues: never }
      | { isMulti: true; selectedValues: string[] }
    );

export const NamespaceSelectorList = ({
  onSelectNamespace,
  isMulti = false,
  selectedValues,
}: NamespaceSelectorListProps) => {
  const { t } = useTranslation();
  const namespace = useNamespace();

  const {
    data: availableNamespaces,
    isLoading,
    isSuccess,
  } = useListNamespaces();

  const hasResults = isSuccess && availableNamespaces?.data.length > 0;

  return (
    <>
      <CommandInput
        placeholder={t("components.breadcrumb.searchPlaceholder")}
      />
      {hasResults && (
        <CommandList className="max-h-[278px]">
          <CommandEmpty>{t("components.breadcrumb.notFound")}</CommandEmpty>
          <CommandGroup>
            {availableNamespaces?.data.map((ns) => (
              <CommandItem
                key={ns.name}
                value={ns.name}
                onSelect={(currentValue: string) => {
                  onSelectNamespace(currentValue);
                }}
              >
                {isMulti ? (
                  <>
                    <Checkbox
                      checked={selectedValues.includes(ns.name)}
                      className="mr-2"
                    />
                    <span>{ns.name}</span>
                  </>
                ) : (
                  <>
                    <Circle
                      className={twMergeClsx(
                        "mr-2 h-2 w-2 fill-current",
                        namespace === ns.name ? "opacity-100" : "opacity-0"
                      )}
                    />
                    <span>{ns.name}</span>
                  </>
                )}
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      )}
      {isLoading && (
        <CommandStaticItem>
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          {t("components.breadcrumb.loading")}
        </CommandStaticItem>
      )}
    </>
  );
};
