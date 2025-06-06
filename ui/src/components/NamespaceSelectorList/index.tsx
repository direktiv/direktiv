import { Check, Circle, Loader2, Square } from "lucide-react";
import {
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandStaticItem,
} from "~/design/Command";

import InfoTooltip from "../NamespaceEdit/InfoTooltip";
import { twMergeClsx } from "~/util/helpers";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

type MultiSelectorProps = { isMultiSelect: true; selectedValues: string[] };
type SingleSelectorProps = { isMultiSelect?: false; selectedValues?: never };

type NamespaceSelectorListProps = {
  onSelectNamespace: (value: string) => void;
} & (MultiSelectorProps | SingleSelectorProps);

export const NamespaceSelectorList = ({
  onSelectNamespace,
  isMultiSelect = false,
  selectedValues = [],
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
                onSelect={(value) => onSelectNamespace(value)}
              >
                {isMultiSelect ? (
                  <>
                    {selectedValues.includes(ns.name) ? (
                      <Check className="mr-2 size-5" />
                    ) : (
                      <Square className="mr-2 size-5" />
                    )}
                    <span>{ns.name}</span>
                  </>
                ) : (
                  <>
                    <Circle
                      className={twMergeClsx(
                        "mr-2 size-2 fill-current",
                        namespace === ns.name ? "opacity-100" : "opacity-0"
                      )}
                    />
                    <span className="flex w-full items-center gap-1">
                      {ns.name}
                      {ns.isSystemNamespace && (
                        <InfoTooltip>
                          {t(
                            "components.namespaceSelector.systemNamespaceTooltip"
                          )}
                        </InfoTooltip>
                      )}
                    </span>
                  </>
                )}
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      )}
      {isLoading && (
        <CommandStaticItem>
          <Loader2 className="mr-2 size-4 animate-spin" />
          {t("components.breadcrumb.loading")}
        </CommandStaticItem>
      )}
    </>
  );
};
