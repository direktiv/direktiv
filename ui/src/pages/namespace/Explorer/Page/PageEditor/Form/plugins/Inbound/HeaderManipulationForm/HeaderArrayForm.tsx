import { ArrayForm } from "~/components/Form/Array";
import { ControllerRenderProps } from "react-hook-form";
import { HeaderManipulationFormSchemaType } from "../../../../schema/plugins/inbound/headerManipulation";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type HeaderArrayFormProps = {
  field:
    | ControllerRenderProps<
        HeaderManipulationFormSchemaType,
        "configuration.headers_to_add"
      >
    | ControllerRenderProps<
        HeaderManipulationFormSchemaType,
        "configuration.headers_to_modify"
      >
    | ControllerRenderProps<
        HeaderManipulationFormSchemaType,
        "configuration.headers_to_remove"
      >;
};

export const HeaderArrayForm = ({ field }: HeaderArrayFormProps) => {
  const { t } = useTranslation();
  const isHeadersToRemove = field.name === "configuration.headers_to_remove";
  const emptyItem = isHeadersToRemove ? { name: "" } : { name: "", value: "" };
  return (
    <div className="grid gap-5">
      <ArrayForm
        defaultValue={field.value || []}
        onChange={field.onChange}
        emptyItem={emptyItem}
        itemIsValid={(item) =>
          !!item && Object.values(item).every((v) => v !== "")
        }
        renderItem={({ value: objectValue, setValue, handleKeyDown }) => (
          <>
            {Object.entries(objectValue).map(([key, value]) => {
              const typedKey = key as keyof typeof objectValue;
              return (
                <Input
                  key={key}
                  data-testid={`env-${typedKey}`}
                  placeholder={t(
                    `pages.explorer.endpoint.editor.form.plugins.inbound.headerManipulation.${typedKey}Placeholder`
                  )}
                  value={value}
                  onKeyDown={handleKeyDown}
                  onChange={(e) => {
                    const newObject = {
                      ...objectValue,
                      [key]: e.target.value,
                    };
                    setValue(newObject);
                  }}
                />
              );
            })}
          </>
        )}
      />
    </div>
  );
};
