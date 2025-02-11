import { ArrayForm } from "~/components/Form/Array";
import { Card } from "~/design/Card";
import Input from "~/design/Input";
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
  return (
    <fieldset className="flex items-center gap-5">
      <label className="w-[90px] text-right text-[14px]">
        {t("pages.permissions.oidcGroupSelector.label")}
      </label>
      <Card
        className="max-h-[200px] w-full overflow-scroll p-5 grid grid-cols-2 gap-5"
        noShadow
      >
        <ArrayForm
          defaultValue={oidcGroups}
          onChange={onChange}
          emptyItem=""
          itemIsValid={(item) => {
            if (!item) return false;
            if (item.includes(",")) return false;
            if (item.includes(" ")) return false;
            return true;
          }}
          renderItem={({ value, setValue, handleKeyDown }) => (
            <Input
              placeholder={t("pages.permissions.oidcGroupSelector.placeholder")}
              className="basis-full"
              value={value}
              onKeyDown={handleKeyDown}
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
