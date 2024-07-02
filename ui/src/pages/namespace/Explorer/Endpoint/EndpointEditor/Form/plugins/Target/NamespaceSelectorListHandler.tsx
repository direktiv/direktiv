import { Command } from "~/design/Command";
import { NamespaceSelectorList } from "~/components/Breadcrumb/NamespaceSelectorList";

type Props = {
  value: string[];
  onValueChange: (value: string[]) => void;
  id?: string;
};

export const NamespaceSelectorListHandler = ({
  value: selectedNamespaces,
  onValueChange,
  id,
}: Props) => (
  <Command id={id}>
    <NamespaceSelectorList
      onSelectNamespace={(value) => {
        if (selectedNamespaces.includes(value)) {
          return onValueChange(
            selectedNamespaces.filter((item: string) => item !== value)
          );
        }
        onValueChange([...selectedNamespaces, value]);
      }}
      isMulti={true}
      selectedValues={selectedNamespaces || []}
    />
  </Command>
);
