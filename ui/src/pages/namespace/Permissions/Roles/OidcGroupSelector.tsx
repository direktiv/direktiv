import { ArrayForm } from "~/components/Form/Array";
import { Card } from "~/design/Card";
import Input from "~/design/Input";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type PermisionsSelectorProps = {
  oidcGroups: string[];
  onChange: (oidcGroups: string[]) => void;
};

const OidcGroupSelector = ({
  oidcGroups,
  onChange,
}: PermisionsSelectorProps) => {
  const { t } = useTranslation();
  const [value, setValue] = useState(oidcGroups);

  const handleChange = (newGroups: string[]) => {
    setValue(newGroups);
    onChange(newGroups);
  };

  return (
    <fieldset className="flex items-center gap-5">
      <label className="w-[120px] text-right text-[14px]">
        {t("pages.permissions.oidcGroupSelector.label")}
      </label>
      <Card
        className="grid max-h-[200px] w-full grid-cols-2 gap-5 overflow-scroll p-5"
        noShadow
      >
        <ArrayForm
          defaultValue={value}
          onChange={handleChange}
          emptyItem=""
          renderItem={({ value, setValue }) => (
            <Input
              placeholder={t("pages.permissions.oidcGroupSelector.placeholder")}
              className="basis-full"
              value={value}
              onChange={(e) => {
                const newValue = e.target.value;
                setValue(newValue);
              }}
            />
          )}
        />
      </Card>
    </fieldset>
  );
};

export default OidcGroupSelector;
